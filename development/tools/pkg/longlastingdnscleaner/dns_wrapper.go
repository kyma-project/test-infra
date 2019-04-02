package dnscleaner

import (
	"context"
	"errors"
	"net/http"

	dns "google.golang.org/api/dns/v1"
)

// DNSAPIWrapper abstracts GCP DNS Service API
type DNSAPIWrapper struct {
	Context context.Context
	Service *dns.Service
}

// LookupDNSEntry delegates to DNS.Service.ResourceRecordSets.List(project, zone, name, address, recordType, recordTTL) function
func (daw *DNSAPIWrapper) LookupDNSEntry(project, zone, name, address, recordType string, recordTTL int64) (*dns.ResourceRecordSet, bool, error) {
	listResp, listErr := daw.Service.ResourceRecordSets.List(project, zone).Name(name).Do()
	if listErr != nil {
		return nil, false, listErr
	}
	if listResp.HTTPStatusCode == http.StatusTooManyRequests {
		return nil, true, errors.New("Quota reached")
	}

	for _, rrs := range listResp.Rrsets {
		if rrs.Type == recordType && rrs.Ttl == recordTTL {
			for _, rrsdata := range rrs.Rrdatas {
				if rrsdata == address {
					return rrs, true, nil
				}
			}
		}
	}

	return nil, false, errors.New("Could not locate DNS record")
}

// RemoveDNSEntry delegates to DNS.Service.Changes.Create(project, zone, *record) function
func (daw *DNSAPIWrapper) RemoveDNSEntry(project, zone string, record *dns.ResourceRecordSet) (bool, error) {
	proposedChange := &dns.Change{}
	proposedChange.Deletions = append(proposedChange.Deletions, record)

	changeResp, changeErr := daw.Service.Changes.Create(project, zone, proposedChange).Do()
	retryStatus := (changeResp.HTTPStatusCode != http.StatusAccepted)
	if changeErr != nil {
		return retryStatus, changeErr
	}
	if changeResp.HTTPStatusCode == http.StatusTooManyRequests {
		return retryStatus, errors.New("Quota reached")
	}

	return retryStatus, nil
}
