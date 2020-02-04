package dnscollector

import (
	"regexp"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

//go:generate mockery -name=ComputeAPI -output=automock -outpkg=automock -case=underscore

//ComputeAPI abstracts over google Compute API
type ComputeAPI interface {
	LookupIPAddresses(project string, region string) ([]*compute.Address, error)
	DeleteIPAddress(project string, region string, address string) error
}

//go:generate mockery -name=DNSAPI -output=automock -outpkg=automock -case=underscore

//DNSAPI abstracts over google DNS API
type DNSAPI interface {
	LookupDNSRecords(project string, managedZone string) ([]*dns.ResourceRecordSet, error)
	DeleteDNSRecord(project string, managedZone string, record *dns.ResourceRecordSet) error
}

//Collector can find and delete IP addresses and DNS records in a GCP project
type Collector struct {
	computeAPI   ComputeAPI
	dnsAPI       DNSAPI
	shouldRemove IPAddressRemovalPredicate
}

//IPAddressRemovalPredicate returns true if IP Address matches removal criteria
type IPAddressRemovalPredicate func(*compute.Address) (bool, error)

// DefaultIPAddressRemovalPredicate returns the default IPAddressRemovalPredicate
// Matching criteria are:
// - name matches one of provided regular expressions.
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
		ageMatches = ipAddressThreshold > 0

		return nameMatches && ageMatches, nil
	}
}

//Wrapper to carry region info along with compute.Address
type addressWrapper struct {
	data   *compute.Address
	region string
}

//New returns an new instance of the Collector
func New(computeAPI ComputeAPI, dnsAPI DNSAPI, removalPredicate IPAddressRemovalPredicate) *Collector {
	return &Collector{computeAPI, dnsAPI, removalPredicate}
}

// Run executes the collection process
func (gc *Collector) Run(project string, managedZone string, regions []string, makeChanges bool) (allSucceeded bool, err error) {

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	common.Shout("Looking for matching IP Addresses and DNS Records in project: \"%s\" and zone: \"%s\" ...", project, managedZone)

	matchingIPs, allSucceeded := gc.listIPs(project, regions)

	if len(matchingIPs) > 0 {
		log.Infof("%sFound %d matching IP Addresses", msgPrefix, len(matchingIPs))
		common.Shout("Removing matching IP Addresses / DNS Records ...")
	} else {
		log.Infof("%sFound no IP Addresses to delete", msgPrefix)
		return true, nil
	}

	allDNSRecords, err := gc.listDNSRecords(project, managedZone)

	if err != nil {
		log.Errorf("listing DNS Records for project \"%s\" and zone \"%s\", %v", project, managedZone, err)
		return false, err
	}

	for _, ipAddress := range matchingIPs {

		associatedRecords := findAssociatedRecords(ipAddress.data.Address, allDNSRecords)

		log.Infof("Processing IP Adress: %s, name: \"%s\", region: \"%s\", %d associated DNS record(s)", ipAddress.data.Address, ipAddress.data.Name, ipAddress.region, len(associatedRecords))

		allDNSRecordsRemoved := true
		for _, dnsRecord := range associatedRecords {
			if makeChanges {
				err = gc.dnsAPI.DeleteDNSRecord(project, managedZone, dnsRecord)
				if err != nil {
					log.Errorf("deleting DNS Records \"%s\": %v", dnsRecord.Name, err)
					allDNSRecordsRemoved = false
					continue
				}
			}
			log.Infof("%sRequested DNS record delete: \"%s\". Zone: \"%s\"", msgPrefix, dnsRecord.Name, managedZone)
		}

		if !allDNSRecordsRemoved {
			allSucceeded = false
			continue //Do NOT remove IP Address yet, next Run might succeed
		}

		if makeChanges {
			err = gc.computeAPI.DeleteIPAddress(project, ipAddress.region, ipAddress.data.Name)
			if err != nil {
				log.Errorf("deleting IP Address \"%s\": %v", ipAddress.data.Address, err)
				allSucceeded = false
				continue
			}
		}
		log.Infof("%sRequested IP Address delete: \"%s\". CreationTimestamp: \"%s\"", msgPrefix, ipAddress.data.Address, ipAddress.data.CreationTimestamp)
	}

	return allSucceeded, nil
}

//List IP Addresses in all regions. It's a "best effort" implementation - continues reading in case of errors.
func (gc *Collector) listIPs(project string, regions []string) (res []*addressWrapper, allSucceeded bool) {
	allSucceeded = true
	res = []*addressWrapper{}

	for _, region := range regions {
		ipAddresses, allProcessed, err := gc.listRegionIPs(project, region)
		if err != nil {
			log.Errorf("Could not list IP Addresses in Region \"%s\": %v", region, err)
			allSucceeded = false
			continue
		}

		if !allProcessed {
			allSucceeded = false
		}

		res = append(res, ipAddresses...)
	}

	return res, allSucceeded
}

//Lists matching IP Addresses in given region
func (gc *Collector) listRegionIPs(project, region string) (res []*addressWrapper, allSucceeded bool, err error) {
	res = []*addressWrapper{}

	rawIPs, err := gc.computeAPI.LookupIPAddresses(project, region)
	if err != nil {
		log.Errorf("Could not list IP Addresses in Region \"%s\": %v", region, err)
		return nil, false, err
	}

	allSucceeded = true
	for _, rawIP := range rawIPs {
		addressMatches, err := gc.shouldRemove(rawIP)

		if err != nil {
			log.Errorf("During verification of IP Address %s (\"%s\"): %v", rawIP.Address, rawIP.Name, err)
			allSucceeded = false
			continue
		}

		if addressMatches {
			ipAddress := wrapAddress(rawIP, region)
			res = append(res, ipAddress)
		}
	}

	return res, allSucceeded, nil
}

func (gc *Collector) listDNSRecords(project, managedZone string) ([]*dns.ResourceRecordSet, error) {
	rawDNSRecords, err := gc.dnsAPI.LookupDNSRecords(project, managedZone)

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

func findAssociatedRecords(ipAddress string, dnsRecords []*dns.ResourceRecordSet) []*dns.ResourceRecordSet {
	res := []*dns.ResourceRecordSet{}
	for _, rec := range dnsRecords {
		if rec.Rrdatas[0] == ipAddress {
			res = append(res, rec)
		}
	}

	return res
}

func wrapAddress(address *compute.Address, region string) *addressWrapper {
	return &addressWrapper{address, region}
}
