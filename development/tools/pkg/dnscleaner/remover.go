package dnscleaner

import (
	"context"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	dns "google.golang.org/api/dns/v1"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=DNSAPI -output=automock -outpkg=automock -case=underscore

// DNSAPI abstracts access to DNS API in GCP
type DNSAPI interface {
	RemoveDNSEntry(ctx context.Context, project, zone string, record *dns.ResourceRecordSet) error
	LookupDNSEntry(ctx context.Context, project, zone, name, address, recordType string, recordTTL int64) (*dns.ResourceRecordSet, error)
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
func (der *DNSEntryRemover) Run(project, zone, dnsName, dnsAddress, recordType string, recordTTL int64) error {
	common.Shout("Trying to retrieve DNS entry with name \"%s\" in project \"%s\", available in zone \"%s\" with Address: \"%s\"", dnsName, project, zone, dnsAddress)

	ctx := context.Background()

	backoff := der.backoff
	var getErr error
	var recordSet *dns.ResourceRecordSet
	for attempt := uint(0); attempt < der.maxAttempts; attempt++ {
		entry, lookupErr := der.dnsAPI.LookupDNSEntry(ctx, project, zone, dnsName, dnsAddress, recordType, recordTTL)
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
		return getErr
	}

	common.Shout("Trying to delete DNS retrieved record set with name \"%s\" in project \"%s\", available in zone \"%s\" with Address: \"%s\"", dnsName, project, zone, dnsAddress)

	var msgPrefix string
	if !der.makeChanges {
		msgPrefix = "[DRY RUN] "
	}

	var delErr error
	backoff = der.backoff
	if der.makeChanges {
		for attempt := uint(0); attempt < der.maxAttempts; attempt++ {
			removalErr := der.dnsAPI.RemoveDNSEntry(ctx, project, zone, recordSet)
			if removalErr == nil {
				delErr = nil
				break
			}
			if removalErr.Error() == dnsDeletionFailed {
				delErr = removalErr
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
		log.Errorf("Could not delete DNS entry with name \"%s\" in zone \"%s\" of project \"%s\", got error: %s", dnsName, zone, project, delErr.Error())
	} else {
		log.Infof("%sRequested deletion of DNS entry with name \"%s\" in zone \"%s\" of project \"%s\"", msgPrefix, dnsName, zone, project)
	}

	return delErr
}
