package dnscleaner

import (
	"fmt"
	"log"
	"strings"
	"time"

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

//ComputeAPI ???
type ComputeAPI interface {
	lookupRegions(project string, pattern string) ([]string, error)
	lookupIPAddresses(project string, region string) ([]*compute.Address, error)
	deleteIPAddress(project string, region string, address string)
}

//DNSAPI ???
type DNSAPI interface {
	lookupDNSRecords(project string, zone string) ([]*dns.ResourceRecordSet, error)
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

//Run runs the logic
func (cleaner *Cleaner) Run(project string, zone string, dryRun bool) {
	regions, err := cleaner.computeAPI.lookupRegions(project, "europe-*")
	if err != nil {
		log.Fatalf("Could not list Zones: %v", err)
	}

	rawRecords, err := cleaner.dnsAPI.lookupDNSRecords(project, zone)
	if err != nil {
		log.Fatalf("Could not list DNSRecords: %v", err)
	}
	parsedRecords := extractRecords(rawRecords, "gkeint")
	fmt.Printf("%s\n\n", parsedRecords)

	var parsedAddresses []IPAddress
	for _, region := range regions {
		rawIP, err := cleaner.computeAPI.lookupIPAddresses(project, region)
		if err != nil {
			log.Fatalf("Could not list IPAddress: %v", err)
		}
		parsedAddresses = append(parsedAddresses, extractIPAddresses(rawIP, region)...)
	}

	cleaner.purge(project, parsedRecords, rawRecords, parsedAddresses, dryRun)

}

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

func (cleaner *Cleaner) purge(project string, records []DNSRecord, rawRecords []*dns.ResourceRecordSet, ips []IPAddress, dryRun bool) {
	for _, record := range records {
		for _, recordIP := range record.records {
			ad := findIPAddress(recordIP, ips)
			fmt.Printf("---> Attempting to remove %s in %s\n", ad, recordIP)
			if !dryRun {
				cleaner.computeAPI.deleteIPAddress(project, findIPRegion(ad, ips), ad)
				time.Sleep(2 * time.Second)
			}
		}
	}
}
