package dnscleaner

import (
	"context"

	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

//ComputeServiceWrapper A wrapper for compute API service connections.
type ComputeServiceWrapper struct {
	Context context.Context
	Compute *compute.Service
}

//DNSServiceWrapper A wrapper for dns API service connections.
type DNSServiceWrapper struct {
	Context context.Context
	DNS     *dns.Service
}

func (csw *ComputeServiceWrapper) lookupRegions(project, pattern string) ([]string, error) {
	call := csw.Compute.Regions.List(project)
	if pattern != "" {
		call = call.Filter("name: " + pattern)
	}

	var regions []string
	f := func(page *compute.RegionList) error {
		for _, v := range page.Items {
			regions = append(regions, v.Name)
		}
		return nil
	}

	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return regions, nil
}

func (dsw *DNSServiceWrapper) lookupDNSRecords(project string, zone string) ([]*dns.ResourceRecordSet, error) {
	call := dsw.DNS.ResourceRecordSets.List(project, zone)

	var items = []*dns.ResourceRecordSet{}
	f := func(page *dns.ResourceRecordSetsListResponse) error {
		for _, v := range page.Rrsets {
			items = append(items, v)
		}
		return nil
	}

	if err := call.Pages(dsw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (csw *ComputeServiceWrapper) lookupIPAddresses(project string, region string) ([]*compute.Address, error) {
	var items = []*compute.Address{}
	call := csw.Compute.Addresses.List(project, region)
	call = call.Filter("status: RESERVED")
	//2018-11-09T05:01:51.510-08:00
	//call = call.Filter("creationTimestamp > 11") <- probably can't be done. Filtering in memory.
	f := func(page *compute.AddressList) error {
		for _, v := range page.Items {
			items = append(items, v)
		}
		return nil
	}

	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (csw *ComputeServiceWrapper) deleteIPAddress(project string, region string, address string) error {
	/*
		_, err := csw.Compute.Addresses.Delete(project, region, address).Do()
		if err != nil {
			return err
		}
	*/
	return nil
}

func (dsw *DNSServiceWrapper) deleteDNSRecord(project string, zone string, record *dns.ResourceRecordSet) error {
	/*
		request := &dns.Change{
			Deletions: []*dns.ResourceRecordSet{record},
		}
		_, err := dsw.DNS.Changes.Create(project, zone, request).Do()
		if err != nil {
			return err
		}
	*/
	return nil
}
