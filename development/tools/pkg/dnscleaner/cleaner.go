package dnscleaner

import (
	"strings"
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
type IPAddress struct {
	name              string
	address           string
	region            string
	creationTimestamp string
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
	deleteDNSRecord(project string, zone string, record *dns.ResourceRecordSet) error
}

//Cleaner ???
type Cleaner struct {
	computeAPI ComputeAPI
	dnsAPI     DNSAPI
}

//NewCleaner ???
func NewCleaner(computeAPI ComputeAPI, dnsAPI DNSAPI) *Cleaner {
	return &Cleaner{computeAPI, dnsAPI}
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

		addressMatches, err := matchIPAddress(rawIP)

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

func matchIPAddress(rawIP *compute.Address) (bool, error) {

	//TODO: Parametrize
	var minAgeInHours = 10

	ipCreationTime, err := time.Parse(time.RFC3339, rawIP.CreationTimestamp)
	if err != nil {
		log.Errorf("Error while parsing CreationTimestamp: \"%s\" for IP Address: %s", rawIP.CreationTimestamp, rawIP.Address)
		return false, err
	}

	ipAddressThreshold := time.Since(ipCreationTime).Hours() - float64(minAgeInHours)
	log.Infof("ipAddressThreshold: %v", ipAddressThreshold)
	ageMatches := ipAddressThreshold > 0

	if ageMatches {
		return true, nil
	}

	return false, nil
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
func (gc *Cleaner) Run(project string, zone string, dryRun bool) (allSucceeded bool, err error) {

	common.Shout("Looking for matching IP Addresses and DNS Records in \"%s\" project and \"%s\" zone...", project, zone)

	matchingIPs, allSucceeded, err := gc.listIPs(project)
	if err != nil {
		return
	}

	var msgPrefix string
	if dryRun {
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

		dnsRecords := findAssociatedRecords(ipAddress.address, dnsRecords)

		log.Infof("Processing IP Adress: %s, name: \"%s\", %d associated DNS record(s)", ipAddress.address, ipAddress.name, len(dnsRecords))

		/*
			for _, dnsRecord := range gd.DNSRecords {
				err = gc.dnsAPI.deleteDNSRecord("", "", nil)
				if err != nil {
					log.Errorf("deleting DNS Records %s: %#v", "DNS", err)
					allSucceeded = false
				} else {
					log.Infof("%sRequested DNS record delete: \"%s\". Project \"%s\", zone \"%s\"", msgPrefix, dnsRecord.name, project, zone)
				}
			}
		*/

		err = gc.computeAPI.deleteIPAddress("", "", "")
		if err != nil {
			log.Errorf("deleting IP Address %s: %#v", "IP_ADDRESS", err)
			allSucceeded = false
		} else {
			log.Infof("%sRequested IP Address delete: \"%s\". Project \"%s\", zone \"%s\", creationTimestamp: \"%s\"", msgPrefix, ipAddress.address, project, zone, ipAddress.creationTimestamp)
		}
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

func extractIPAddress(raw *compute.Address, region string) *IPAddress {
	return &IPAddress{raw.Name, raw.Address, region, raw.CreationTimestamp}
}

func extractRecords(records []*dns.ResourceRecordSet, key string) []DNSRecord {
	var items []DNSRecord
	for _, record := range records {
		if strings.Contains(record.Name, key) {
			items = append(items, DNSRecord{record.Name, record.Rrdatas[0]})
		}
	}
	return items
}
