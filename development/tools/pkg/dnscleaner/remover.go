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
	RemoveDNSEntry(project, zone string, record *dns.ResourceRecordSet) error
	LookupDNSEntry(project, zone, name, address, recordType string, recordTTL int64) (*dns.ResourceRecordSet, error)
}

// DNSEntryRemover deletes IPs provisioned by gke-long-lasting prow jobs.
type DNSEntryRemover struct {
	dnsAPI      DNSAPI
	maxAttempts uint
	backoff     uint
	makeChanges bool
}

// New returns a new instance of dnsEntryRemover
func New(dnsAPI DNSAPI, maxAttempts, backoff uint, makeChanges bool) *DNSEntryRemover {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &DNSEntryRemover{dnsAPI, maxAttempts, backoff, makeChanges}
}

// Run executes dns removal process for specified dns record-set
func (der *DNSEntryRemover) Run(project, zone, dnsName, dnsAddress, recordType string, recordTTL int64) (bool, error) {
	common.Shout("Trying to retrieve DNS entry with name \"%s\" in project \"%s\", available in zone \"%s\" with Address: \"%s\"", dnsName, project, zone, dnsAddress)

	backoff := der.backoff
	var getErr error
	var recordSet *dns.ResourceRecordSet
	for attempt := uint(0); attempt < der.maxAttempts; attempt = attempt + 1 {
		entry, lookupErr := der.dnsAPI.LookupDNSEntry(project, zone, dnsName, dnsAddress, recordType, recordTTL)
		if entry != nil {
			recordSet = entry
			break
		}
		if attempt < der.maxAttempts {
			time.Sleep(time.Duration(backoff) * time.Second)
			backoff = backoff * 2
			getErr = lookupErr
		}
	}
	if recordSet == nil {
		return false, getErr
	}

	common.Shout("Trying to delete DNS retrieved record set with name \"%s\" in project \"%s\", available in zone \"%s\" with Address: \"%s\"", dnsName, project, zone, dnsAddress)

	var msgPrefix string
	if !der.makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	var delErr error
	succeeded := true
	backoff = der.backoff
	if der.makeChanges {
		for attempt := uint(0); attempt < der.maxAttempts; attempt = attempt + 1 {
			removalErr := der.dnsAPI.RemoveDNSEntry(project, zone, recordSet)
			if removalErr == nil {
				delErr = nil
				break
			}
			if attempt < der.maxAttempts {
				time.Sleep(time.Duration(backoff) * time.Second)
				backoff = backoff * 2
				delErr = removalErr
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
