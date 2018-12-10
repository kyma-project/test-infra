package dnscleaner

import (
	"fmt"
	"log"
	"strings"

	dns "google.golang.org/api/dns/v1"
)

//DNSRecord Simplified DNS object.
type DNSRecord struct {
	name    string
	records []string
}

//ComputeAPI ???
type ComputeAPI interface {
	lookupZones(project string, pattern string) ([]string, error)
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
func (cleaner *Cleaner) Run(project string, zone string) {
	zones, err := cleaner.computeAPI.lookupZones(project, "europe-*")
	if err != nil {
		log.Fatalf("Could not list Zones: %v", err)
	}
	fmt.Printf("%s\n\n", zones)
	rawRecords, err := cleaner.dnsAPI.lookupDNSRecords(project, zone)
	parsedRecords := extractRecords(rawRecords, "gkeint")
	fmt.Printf("%s\n", parsedRecords)
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
