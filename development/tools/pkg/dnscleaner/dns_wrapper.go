package dnscleaner

import (
	"context"
	"errors"

	dns "google.golang.org/api/dns/v1"
)

// DNSAPIWrapper abstracts GCP DNS Service API
type DNSAPIWrapper struct {
	Context context.Context
	Service *dns.Service
}

// LookupDNSEntry delegates to DNS.Service.ResourceRecordSets.List(project, zone, name, address, recordType, recordTTL) function
func (daw *DNSAPIWrapper) LookupDNSEntry(project, zone, name, address, recordType string, recordTTL int64) (*dns.ResourceRecordSet, error) {
	listResp, listErr := daw.Service.ResourceRecordSets.List(project, zone).Name(name).Context(daw.Context).Do()
	if listErr != nil {
		return nil, listErr
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
func (daw *DNSAPIWrapper) RemoveDNSEntry(project, zone string, record *dns.ResourceRecordSet) error {
	proposedChange := &dns.Change{}
	proposedChange.Deletions = append(proposedChange.Deletions, record)

	_, changeErr := daw.Service.Changes.Create(project, zone, proposedChange).Context(daw.Context).Do()
	if changeErr != nil {
		return changeErr
	}

	return nil
}
