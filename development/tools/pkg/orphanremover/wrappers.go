package orphanremover

import (
	"context"
	"log"

	compute "google.golang.org/api/compute/v1"
)

// ComputeServiceWrapper A wrapper for compute API service connections.
type ComputeServiceWrapper struct {
	Context context.Context
	Compute *compute.Service
}

// DeleteHTTPProxy Delete an httpProxy object
func (csw *ComputeServiceWrapper) DeleteHTTPProxy(project string, httpProxy string) {
	_, err := csw.Compute.TargetHttpProxies.Delete(project, httpProxy).Do()
	if err != nil {
		log.Print(err)
	}
}

// DeleteURLMap Delte an URLMap object
func (csw *ComputeServiceWrapper) DeleteURLMap(project string, urlMap string) {
	_, err := csw.Compute.UrlMaps.Delete(project, urlMap).Do()
	if err != nil {
		log.Print(err)
	}
}

// DeleteBackendService ???
func (csw *ComputeServiceWrapper) DeleteBackendService(project string, backendService string) {
	_, err := csw.Compute.BackendServices.Delete(project, backendService).Do()
	if err != nil {
		log.Print(err)
	}
}

// DeleteInstanceGroup ???
func (csw *ComputeServiceWrapper) DeleteInstanceGroup(project string, zone string, instanceGroup string) {
	_, err := csw.Compute.InstanceGroups.Delete(project, zone, instanceGroup).Do()
	if err != nil {
		log.Print(err)
	}
}

// DeleteHealthChecks ???
func (csw *ComputeServiceWrapper) DeleteHealthChecks(project string, names []string) {
	for _, check := range names {
		_, err := csw.Compute.HttpHealthChecks.Delete(project, check).Do()
		if err != nil {
			log.Print(err)
		}
	}
}

// DeleteForwardingRule ???
func (csw *ComputeServiceWrapper) DeleteForwardingRule(project string, name string, region string) {
	_, err := csw.Compute.ForwardingRules.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

// DeleteGlobalForwardingRule ???
func (csw *ComputeServiceWrapper) DeleteGlobalForwardingRule(project string, name string) {
	_, err := csw.Compute.GlobalForwardingRules.Delete(project, name).Do()
	if err != nil {
		log.Print(err)
	}
}

// DeleteTargetPool ???
func (csw *ComputeServiceWrapper) DeleteTargetPool(project string, name string, region string) {
	_, err := csw.Compute.TargetPools.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

// LookupURLMaps ???
func (csw *ComputeServiceWrapper) LookupURLMaps(project string) ([]*compute.UrlMap, error) {
	call := csw.Compute.UrlMaps.List(project)
	var items []*compute.UrlMap
	f := func(page *compute.UrlMapList) error {
		items = append(items, page.Items...)
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// LookupBackendServices ???
func (csw *ComputeServiceWrapper) LookupBackendServices(project string) ([]*compute.BackendService, error) {
	call := csw.Compute.BackendServices.List(project)
	var items []*compute.BackendService
	f := func(page *compute.BackendServiceList) error {
		items = append(items, page.Items...)
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// LookupInstanceGroup ???
func (csw *ComputeServiceWrapper) LookupInstanceGroup(project string, zone string) ([]string, error) {
	call := csw.Compute.InstanceGroups.List(project, zone)
	call = call.Filter("size: 0")
	var items []string
	f := func(page *compute.InstanceGroupList) error {
		for _, list := range page.Items {
			items = append(items, list.Name)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// LookupTargetPools ???
func (csw *ComputeServiceWrapper) LookupTargetPools(project string) ([]*compute.TargetPool, error) {
	call := csw.Compute.TargetPools.AggregatedList(project)
	var items []*compute.TargetPool
	f := func(page *compute.TargetPoolAggregatedList) error {
		for _, list := range page.Items {
			items = append(items, list.TargetPools...)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// LookupZones ???
func (csw *ComputeServiceWrapper) LookupZones(project, pattern string) ([]string, error) {
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

// LookupHTTPProxy ???
func (csw *ComputeServiceWrapper) LookupHTTPProxy(project string) ([]*compute.TargetHttpProxy, error) {
	call := csw.Compute.TargetHttpProxies.List(project)
	var items []*compute.TargetHttpProxy
	f := func(page *compute.TargetHttpProxyList) error {
		items = append(items, page.Items...)
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// LookupGlobalForwardingRule ???
func (csw *ComputeServiceWrapper) LookupGlobalForwardingRule(project string) ([]*compute.ForwardingRule, error) {
	call := csw.Compute.GlobalForwardingRules.List(project)
	var items []*compute.ForwardingRule
	f := func(page *compute.ForwardingRuleList) error {
		items = append(items, page.Items...)
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

// CheckInstance Verify if instance (vm) of given name exists
func (csw *ComputeServiceWrapper) CheckInstance(project string, zone string, name string) bool {
	_, err := csw.Compute.Instances.Get(project, zone, name).Do()
	return err == nil
}
