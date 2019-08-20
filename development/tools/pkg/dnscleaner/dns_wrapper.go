package dnscleaner

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	dns "google.golang.org/api/dns/v1"
)

const (
	dnsDeletionFailed = "delete failed"
)

// DNSAPIWrapper abstracts GCP DNS Service API
type DNSAPIWrapper struct {
	Service *dns.Service
}

// LookupDNSEntry delegates to DNS.Service.ResourceRecordSets.List(project, zone, name, address, recordType, recordTTL) function
func (daw *DNSAPIWrapper) LookupDNSEntry(ctx context.Context, project, zone, name, address, recordType string, recordTTL int64) (*dns.ResourceRecordSet, error) {
	listResp, listErr := daw.Service.ResourceRecordSets.List(project, zone).Name(name).Context(ctx).Do()
	if listErr != nil {
		return nil, errors.Wrap(listErr, "could not locate DNS entry")
	}

	for _, rrs := range listResp.Rrsets {
		if rrs.Type == recordType && rrs.Ttl == recordTTL {
			for _, rrsdata := range rrs.Rrdatas {
				if rrsdata == address {
					return rrs, nil
				}
			}
		}
	}

	return nil, errors.New("Could not locate DNS record")
}

// RemoveDNSEntry delegates to DNS.Service.Changes.Create(project, zone, *record) function
func (daw *DNSAPIWrapper) RemoveDNSEntry(ctx context.Context, project, zone string, record *dns.ResourceRecordSet) error {
	proposedChange := &dns.Change{}
	proposedChange.Deletions = append(proposedChange.Deletions, record)

	resp, changeErr := daw.Service.Changes.Create(project, zone, proposedChange).Context(ctx).Do()
	if changeErr != nil {
		return errors.Wrap(changeErr, "could not remove DNS entry")
	}
	if resp.HTTPStatusCode > http.StatusAccepted {
		return errors.New(dnsDeletionFailed)
	}

	return nil
}
