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

func (csw *ComputeServiceWrapper) lookupZones(project, pattern string) ([]string, error) {
	call := csw.Compute.Zones.List(project)
	if pattern != "" {
		call = call.Filter("name: " + pattern)
	}

	var zones []string
	f := func(page *compute.ZoneList) error {
		for _, v := range page.Items {
			zones = append(zones, v.Name)
		}
		return nil
	}

	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return zones, nil
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
