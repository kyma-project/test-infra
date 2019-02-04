package dnscollector

import (
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

type addressWrapper struct {
	data   *compute.Address
	region string
}

//ComputeAPI abstracts over google Compute API
type ComputeAPI interface {
	lookupRegions(project string, pattern string) ([]string, error)
	lookupIPAddresses(project string, region string) ([]*compute.Address, error)
	deleteIPAddress(project string, region string, address string) error
}

//DNSAPI abstracts over google DNS API
type DNSAPI interface {
	lookupDNSRecords(project string, managedZone string) ([]*dns.ResourceRecordSet, error)
	deleteDNSRecords(project string, managedZone string, record *dns.ResourceRecordSet) error
}

//Collector ???
type Collector struct {
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
		//log.Infof("ipAddressThreshold: %v", ipAddressThreshold)
		ageMatches = ipAddressThreshold > 0

		return nameMatches && ageMatches, nil
	}
}

//NewCleaner returns an instance of DNSCollector
func NewCleaner(computeAPI ComputeAPI, dnsAPI DNSAPI, removalPredicate IPAddressRemovalPredicate) *Collector {
	return &Collector{computeAPI, dnsAPI, removalPredicate}
}

//List IP Addresses in all regions. It's a "best effort" implementation,
func (gc *Collector) listIPs(project string) (res []*addressWrapper, allSucceeded bool, err error) {
	regions, err := gc.computeAPI.lookupRegions(project, "europe-*")
	if err != nil {
		log.Errorf("Could not list Regions: %v", err)
		return nil, false, err
	}
	allSucceeded = true
	res = []*addressWrapper{}

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
func (gc *Collector) listRegionIPs(project, region string) (res []*addressWrapper, err error) {
	res = []*addressWrapper{}

	rawIPs, err := gc.computeAPI.lookupIPAddresses(project, region)
	if err != nil {
		log.Errorf("Could not list IP Addresses in Region \"%s\": %v", region, err)
		return nil, err
	}

	for _, rawIP := range rawIPs {
		//TODO: remove
		//log.Infof("IP: %#v", rawIP)

		addressMatches, err := gc.shouldRemove(rawIP)

		if err != nil {
			return nil, err
		}

		if addressMatches {
			ipAddress := wrapAddress(rawIP, region)
			res = append(res, ipAddress)
		}
	}

	return res, nil
}

func (gc *Collector) listDNSRecords(project, managedZone string) ([]*dns.ResourceRecordSet, error) {
	rawDNSRecords, err := gc.dnsAPI.lookupDNSRecords(project, managedZone)

	if err != nil {
		return nil, err
	}

	res := []*dns.ResourceRecordSet{}

	for _, rawDNS := range rawDNSRecords {
		if matchDNSRecord(rawDNS) {
			res = append(res, rawDNS)
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
func (gc *Collector) Run(project string, managedZone string, makeChanges bool) (allSucceeded bool, err error) {

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	common.Shout("%sLooking for matching IP Addresses and DNS Records in project: \"%s\" and zone: \"%s\" ...", msgPrefix, project, managedZone)

	matchingIPs, allSucceeded, err := gc.listIPs(project)
	if err != nil {
		return
	}

	if len(matchingIPs) > 0 {
		log.Infof("%sFound %d matching IP Addresses", msgPrefix, len(matchingIPs))
		common.Shout("Removing matching IP Addresses / DNS Records ...")
	} else {
		log.Infof("%sFound no IP Addresses to delete", msgPrefix)
	}

	allDNSRecords, err := gc.listDNSRecords(project, managedZone)

	if err != nil {
		log.Error("listing DNS Records for project \"%s\" and zone \"%s\", %v", project, managedZone, err)
		return false, err
	}

	for _, ipAddress := range matchingIPs {

		associatedRecords := findAssociatedRecords(ipAddress.data.Address, allDNSRecords)

		log.Infof("Processing IP Adress: %s, name: \"%s\", %d associated DNS record(s)", ipAddress.data.Address, ipAddress.data.Name, len(associatedRecords))

		for _, dnsRecord := range associatedRecords {
			if makeChanges {
				err = gc.dnsAPI.deleteDNSRecords(project, managedZone, dnsRecord)
				if err != nil {
					log.Errorf("deleting DNS Records %s: %#v", "DNS", err)
					allSucceeded = false
					continue
				}
			}
			log.Infof("%sRequested DNS record delete: \"%s\". Project \"%s\", zone \"%s\"", msgPrefix, dnsRecord.Name, project, managedZone)
		}

		if makeChanges {
			err = gc.computeAPI.deleteIPAddress("", "", "")
			if err != nil {
				log.Errorf("deleting IP Address %s: %#v", "IP_ADDRESS", err)
				allSucceeded = false
			}
		}
		log.Infof("%sRequested IP Address delete: \"%s\". Project \"%s\", zone \"%s\", creationTimestamp: \"%s\"", msgPrefix, ipAddress.data.Address, project, managedZone, ipAddress.data.CreationTimestamp)
	}

	return allSucceeded, nil
}

func findAssociatedRecords(ipAddress string, dnsRecords []*dns.ResourceRecordSet) []*dns.ResourceRecordSet {
	res := []*dns.ResourceRecordSet{}
	for _, rec := range dnsRecords {
		if rec.Rrdatas[0] == ipAddress {
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

func wrapAddress(address *compute.Address, region string) *addressWrapper {
	return &addressWrapper{address, region}
}
