package dnscleaner

import (
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

//DNSRecord Simplified DNS object.
type DNSRecord struct {
	name      string
	ipAddress string
}

//IPAddress Simplified Address object.
//TODO: Rename to AddressWrapper
type IPAddress struct {
	rawAddress *compute.Address
	region     string
}

type ResourcesToRemove struct {
	IPAddress  *IPAddress
	DNSRecords []*DNSRecord
}

//ComputeAPI ???
type ComputeAPI interface {
	lookupRegions(project string, pattern string) ([]string, error)
	lookupIPAddresses(project string, region string) ([]*compute.Address, error)
	deleteIPAddress(project string, region string, address string) error
}

//DNSAPI ???
type DNSAPI interface {
	lookupDNSRecords(project string, zone string) ([]*dns.ResourceRecordSet, error)
	deleteDNSRecord(project string, zone string, record []*dns.ResourceRecordSet) error
}

//Cleaner ???
type Cleaner struct {
	computeAPI   ComputeAPI
	dnsAPI       DNSAPI
	shouldRemove IPAddressRemovalPredicate
}

// IPAddressRemovalPredicate returns true if IP Address should be deleted (matches removal criteria)
type IPAddressRemovalPredicate func(*compute.Address) (bool, error)

// DefaultIPAddressRemovalPredicate returns the default IPAddressRemovalPredicate
// Matching criteria are:
// - Name matches one of provided regular expressions.
// - CreationTimestamp indicates that it is created more than ageInHours ago.
func DefaultIPAddressRemovalPredicate(addressRegexpList []*regexp.Regexp, minAgeInHours int) IPAddressRemovalPredicate {

	return func(address *compute.Address) (bool, error) {
		nameMatches := false
		ageMatches := false

		for _, r := range addressRegexpList {
			if r.MatchString(address.Name) {
				nameMatches = true
			}
		}

		ipCreationTime, err := time.Parse(time.RFC3339, address.CreationTimestamp)
		if err != nil {
			log.Errorf("Error while parsing CreationTimestamp: \"%s\" for IP Address: %s", address.CreationTimestamp, address.Address)
			return false, err
		}

		ipAddressThreshold := time.Since(ipCreationTime).Hours() - float64(minAgeInHours)
		//TODO: Remove
		log.Infof("ipAddressThreshold: %v", ipAddressThreshold)
		ageMatches = ipAddressThreshold > 0

		return nameMatches && ageMatches, nil
	}
}

//NewCleaner ???
func NewCleaner(computeAPI ComputeAPI, dnsAPI DNSAPI, removalPredicate IPAddressRemovalPredicate) *Cleaner {
	return &Cleaner{computeAPI, dnsAPI, removalPredicate}
}

//List IP Addresses in all regions. It's a "best effort" implementation,
func (gc *Cleaner) listIPs(project string) (res []*IPAddress, allSucceeded bool, err error) {
	regions, err := gc.computeAPI.lookupRegions(project, "europe-*")
	if err != nil {
		log.Errorf("Could not list Regions: %v", err)
		return nil, false, err
	}
	allSucceeded = true
	res = []*IPAddress{}

	for _, regName := range regions {
		ipAddresses, err := gc.listRegionIPs(project, regName)
		if err != nil {
			log.Errorf("Could not list IP Addresses in Region \"%s\": %v", regName, err)
			allSucceeded = false
			continue
		}

		res = append(res, ipAddresses...)
	}

	return res, allSucceeded, nil
}

//Lists matching IP Addresses in given region
func (gc *Cleaner) listRegionIPs(project, region string) (res []*IPAddress, err error) {
	res = []*IPAddress{}

	rawIPs, err := gc.computeAPI.lookupIPAddresses(project, region)
	if err != nil {
		log.Errorf("Could not list IP Addresses in Region \"%s\": %v", region, err)
		return nil, err
	}

	for _, rawIP := range rawIPs {
		//TODO: remove
		log.Infof("IP: %#v", rawIP)

		addressMatches, err := gc.shouldRemove(rawIP)

		if err != nil {
			return nil, err
		}

		if addressMatches {
			ipAddress := extractIPAddress(rawIP, region)
			res = append(res, ipAddress)
		}
	}

	return res, nil
}

func (gc *Cleaner) listDNSRecords(project, zone string) ([]*DNSRecord, error) {
	rawDNSRecords, err := gc.dnsAPI.lookupDNSRecords(project, zone)

	if err != nil {
		return nil, err
	}

	res := []*DNSRecord{}

	for _, rawDNS := range rawDNSRecords {
		if matchDNSRecord(rawDNS) {
			//log.Infof("DNS Record [RAW]: %#v", rawDNS)
			res = append(res, &DNSRecord{rawDNS.Name, rawDNS.Rrdatas[0]})
		}
	}

	return res, nil
}

func matchDNSRecord(rawDNS *dns.ResourceRecordSet) bool {
	typeMatches := rawDNS.Type == "A"
	hasSingleIPAddress := len(rawDNS.Rrdatas) == 1
	return typeMatches && hasSingleIPAddress
}

// Run executes disks garbage collection process
func (gc *Cleaner) Run(project string, zone string, makeChanges bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching IP Addresses and DNS Records in \"%s\" project and \"%s\" zone...", project, zone)

	matchingIPs, allSucceeded, err := gc.listIPs(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	if len(matchingIPs) > 0 {
		log.Infof("%sFound %d matching IP Addresses", msgPrefix, len(matchingIPs))
		common.Shout("Removing matching IP Addresses / DNS Records ...")
	} else {
		log.Infof("%sFound no IP Addresses to delete", msgPrefix)
	}

	dnsRecords, err := gc.listDNSRecords(project, zone)

	if err != nil {
		log.Error("listing DNS Records for project \"%s\" and zone \"%s\", %v", project, zone, err)
		return false, err
	}

	for _, ipAddress := range matchingIPs {

		dnsRecords := findAssociatedRecords(ipAddress.rawAddress.Address, dnsRecords)

		log.Infof("Processing IP Adress: %s, name: \"%s\", %d associated DNS record(s)", ipAddress.rawAddress.Address, ipAddress.rawAddress.Name, len(dnsRecords))

		for _, dnsRecord := range dnsRecords {
			if makeChanges {
				err = gc.dnsAPI.deleteDNSRecord("", "", nil)
				if err != nil {
					log.Errorf("deleting DNS Records %s: %#v", "DNS", err)
					allSucceeded = false
					continue
				}
			}
			log.Infof("%sRequested DNS record delete: \"%s\". Project \"%s\", zone \"%s\"", msgPrefix, dnsRecord.name, project, zone)

		}

		if makeChanges {
			err = gc.computeAPI.deleteIPAddress("", "", "")
			if err != nil {
				log.Errorf("deleting IP Address %s: %#v", "IP_ADDRESS", err)
				allSucceeded = false
			}
		}
		log.Infof("%sRequested IP Address delete: \"%s\". Project \"%s\", zone \"%s\", creationTimestamp: \"%s\"", msgPrefix, ipAddress.rawAddress.Address, project, zone, ipAddress.rawAddress.CreationTimestamp)
	}

	return allSucceeded, nil
}

func findAssociatedRecords(ipAddress string, dnsRecords []*DNSRecord) []*DNSRecord {
	res := []*DNSRecord{}
	for _, rec := range dnsRecords {
		if rec.ipAddress == ipAddress {
			res = append(res, rec)
		}
	}

	return res
}

/*
Found 12 matching IP Addresses with 10 associated DNS records
SHOUT: Removing matching objects...
Processing IP Address: 10.10.1.3 (2 associated DNS Record)
Requested DNS Record delete: "*.gkeint-pr-123.build.kyma.io => 10.10.1.3"
Requested DNS Record delete: "*.gkeint-pr-124.build.kyma.io => 10.10.1.3"
Requested IP Address delete: 10.10.1.3
*/

func countDNSrecords(list []*ResourcesToRemove) int {
	var count = 0
	for _, toRemove := range list {
		count = count + len(toRemove.DNSRecords)
	}

	return count
}

func extractIPAddress(rawAddress *compute.Address, region string) *IPAddress {
	return &IPAddress{rawAddress, region}
}
