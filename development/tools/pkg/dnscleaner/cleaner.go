package dnscleaner

import (
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

//DNSRecord Simplified DNS object.
type DNSRecord struct {
	name    string
	records []string
}

//IPAddress Simplified Address object.
type IPAddress struct {
	name    string
	address string
	region  string
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

func (gc *Cleaner) list() ([]*ResourcesToRemove, error) {
	return nil, nil
}

// Run executes disks garbage collection process
func (gc *Cleaner) Run(project string, zone string, dryRun bool) (allSucceeded bool, err error) {

	//makeChanges := !dryRun
	common.Shout("Looking for matching IP Addresses and DNS Records in \"%s\" project and \"%s\" zone...", project, zone)

	garbageDisks, err := gc.list()
	if err != nil {
		return
	}

	var msgPrefix string
	if dryRun {
		msgPrefix = "[DRY RUN] "
	}

	if len(garbageDisks) > 0 {
		log.Infof("%sFound %d matching disks", msgPrefix, len(garbageDisks))
		common.Shout("Removing matching disks...")
	} else {
		log.Infof("%sFound no disks to delete", msgPrefix)
	}

	allSucceeded = true
	for _, gd := range garbageDisks {

		var err error

		log.Infof("Processing IP Adress: \"%s\": %s (2 associated DNS records)", gd.IPAddress.name, gd.IPAddress.address)

		for _, dnsRecord := range gd.DNSRecords {
			err = gc.dnsAPI.deleteDNSRecord("", "", nil)
			if err != nil {
				log.Errorf("deleting DNS Records %s: %#v", "DNS", err)
				allSucceeded = false
			} else {
				log.Infof("%sRequested DNS record delete: \"%s\". Project \"%s\", zone \"%s\"", msgPrefix, dnsRecord.name, project, zone)
			}
		}

		err = gc.computeAPI.deleteIPAddress("", "", "")
		if err != nil {
			log.Errorf("deleting IP Address %s: %#v", "IP_ADDRESS", err)
			allSucceeded = false
		} else {
			log.Infof("%sRequested IP Address delete: \"%s\". Project \"%s\", zone \"%s\", disk creationTimestamp: \"%s\"", msgPrefix, gd.IPAddress.address, project, zone, "")
		}
	}

	return allSucceeded, nil
}

/*
Found 12 matching IP Addresses with 10 associated DNS entries
SHOUT: Removing matching objects...
Processing IP Address: 10.10.1.3 (2 associated DNS Record)
Requested DNS Record delete: "*.gkeint-pr-123.build.kyma.io => 10.10.1.3"
Requested DNS Record delete: "*.gkeint-pr-124.build.kyma.io => 10.10.1.3"
Requested IP Address delete: 10.10.1.3
*/

func extractIPAddresses(raw []*compute.Address, region string) []IPAddress {
	var parsed []IPAddress
	for _, ip := range raw {
		parsed = append(parsed, IPAddress{ip.Name, ip.Address, region})
	}
	return parsed
}

func extractRecords(records []*dns.ResourceRecordSet, key string) []DNSRecord {
	var items []DNSRecord
	for _, record := range records {
		if strings.Contains(record.Name, key) {
			items = append(items, DNSRecord{record.Name, record.Rrdatas})
		}
	}
	return items
}

func findRecord(name string, rawRecords []*dns.ResourceRecordSet) *dns.ResourceRecordSet {
	for it, record := range rawRecords {
		if record.Name == name {
			return rawRecords[it]
		}
	}
	return nil
}

func findIPAddress(record string, addresses []IPAddress) string {
	for _, ip := range addresses {
		if ip.address == record {
			return ip.name
		}
	}
	return "NOT_FOUND"
}

func findIPRegion(name string, addresses []IPAddress) string {
	for _, ip := range addresses {
		if ip.name == name {
			return ip.region
		}
	}
	return "NOT_FOUND"
}
