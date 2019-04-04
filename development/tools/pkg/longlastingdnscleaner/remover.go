package dnscleaner

import (
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	dns "google.golang.org/api/dns/v1"
)

//go:generate mockery -name=DNSAPI -output=automock -outpkg=automock -case=underscore

// DNSAPI abstracts access to DNS API in GCP
type DNSAPI interface {
	RemoveDNSEntry(project, zone string, record *dns.ResourceRecordSet) (bool, error)
	LookupDNSEntry(project, zone, name, address, recordType string, recordTTL int64) (*dns.ResourceRecordSet, bool, error)
}

// DNSEntryRemover deletes IPs provisioned by gke-long-lasting prow jobs.
type DNSEntryRemover struct {
	dnsAPI DNSAPI
}

// NewDNSEntryRemover returns a new instance of DNSEntryRemover
func NewDNSEntryRemover(dnsAPI DNSAPI) *DNSEntryRemover {
	return &DNSEntryRemover{dnsAPI}
}

// Run executes dns removal process for specified dns record-set
func (der *DNSEntryRemover) Run(project, zone, dnsName, dnsAddress, recordType string, recordTTL int64, maxAttempts, timeout uint, makeChanges bool) (bool, error) {
	common.Shout("Trying to retrieve DNS entry with name \"%s\" in project \"%s\", available in zone \"%s\" with Address: \"%s\"", dnsName, project, zone, dnsAddress)

	var getErr error
	var recordSet *dns.ResourceRecordSet
	attempts := uint(0)
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	for {
		entry, retryable, lookupErr := der.dnsAPI.LookupDNSEntry(project, zone, dnsName, dnsAddress, recordType, recordTTL)
		if entry != nil {
			recordSet = entry
			break
		}
		attempts = attempts + 1
		if attempts < maxAttempts && retryable {
			time.Sleep(time.Duration(timeout) * time.Second)
			timeout = timeout * 2
		} else {
			getErr = lookupErr
			break
		}
	}
	if recordSet == nil {
		return false, getErr
	}

	common.Shout("Trying to delete DNS retrieved record set with name \"%s\" in project \"%s\", available in zone \"%s\" with Address: \"%s\"", dnsName, project, zone, dnsAddress)

	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	var delErr error
	succeeded := true
	attempts = uint(0)
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if makeChanges {
		for {
			retryable, removalErr := der.dnsAPI.RemoveDNSEntry(project, zone, recordSet)
			attempts = attempts + 1
			if attempts < maxAttempts && retryable {
				time.Sleep(time.Duration(timeout) * time.Second)
				timeout = timeout * 2
			} else {
				delErr = removalErr
				break
			}
		}
	}
	if delErr != nil {
		log.Infof("Could not delete DNS entry with name \"%s\" in zone \"%s\" of project \"%s\", got error: %s", dnsName, zone, project, delErr.Error())
		succeeded = false
	} else {
		log.Infof("%sRequested deletion of DNS entry with name \"%s\" in zone \"%s\" of project \"%s\"", msgPrefix, dnsName, zone, project)
	}

	return succeeded, delErr
}
