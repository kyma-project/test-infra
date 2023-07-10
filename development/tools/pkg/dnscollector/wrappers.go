package dnscollector

import (
	"context"

	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

// ComputeServiceWrapper A wrapper for compute API service.
type ComputeServiceWrapper struct {
	Context context.Context
	Compute *compute.Service
}

// DNSServiceWrapper A wrapper for dns API service.
type DNSServiceWrapper struct {
	Context context.Context
	DNS     *dns.Service
}

// LookupIPAddresses delegates to GCP Compute API to find all IP Addresses with a "RESERVED" status in a given region.
func (csw *ComputeServiceWrapper) LookupIPAddresses(project string, region string) ([]*compute.Address, error) {
	var items = []*compute.Address{}
	call := csw.Compute.Addresses.List(project, region)
	call = call.Filter("status: RESERVED")
	f := func(page *compute.AddressList) error {
		items = append(items, page.Items...)
		return nil
	}

	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteIPAddress delegates to GCP API to deletes specified IP Address
func (csw *ComputeServiceWrapper) DeleteIPAddress(project string, region string, address string) error {
	_, err := csw.Compute.Addresses.Delete(project, region, address).Do()
	if err != nil {
		return err
	}
	return nil
}

// LookupDNSRecords delegates to GCP DNS API do find DNS Records in a specified managed zone.
func (dsw *DNSServiceWrapper) LookupDNSRecords(project string, managedZone string) ([]*dns.ResourceRecordSet, error) {
	call := dsw.DNS.ResourceRecordSets.List(project, managedZone)

	var items = []*dns.ResourceRecordSet{}
	f := func(page *dns.ResourceRecordSetsListResponse) error {
		items = append(items, page.Rrsets...)
		return nil
	}

	if err := call.Pages(dsw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteDNSRecord delegates to GCP DNS API to delete specified DNS record
func (dsw *DNSServiceWrapper) DeleteDNSRecord(project string, managedZone string, recordToDelete *dns.ResourceRecordSet) error {
	change := &dns.Change{
		Deletions: []*dns.ResourceRecordSet{recordToDelete},
	}

	_, err := dsw.DNS.Changes.Create(project, managedZone, change).Context(dsw.Context).Do()
	if err != nil {
		return err
	}

	return nil
}
