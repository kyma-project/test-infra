package orphanremover

import (
	"context"
	"log"

	compute "google.golang.org/api/compute/v1"
)

//ComputeServiceWrapper A wrapper for compute API service connections.
type ComputeServiceWrapper struct {
	Context context.Context
	Compute *compute.Service
}

func (csw *ComputeServiceWrapper) deleteHTTPProxy(project string, httpProxy string) {
	_, err := csw.Compute.TargetHttpProxies.Delete(project, httpProxy).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) deleteURLMap(project string, urlMap string) {
	_, err := csw.Compute.UrlMaps.Delete(project, urlMap).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) deleteBackendService(project string, backendService string) {
	_, err := csw.Compute.BackendServices.Delete(project, backendService).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) deleteInstanceGroup(project string, zone string, instanceGroup string) {
	_, err := csw.Compute.InstanceGroups.Delete(project, zone, instanceGroup).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) deleteHealthChecks(project string, names []string) {
	for _, check := range names {
		_, err := csw.Compute.HttpHealthChecks.Delete(project, check).Do()
		if err != nil {
			log.Print(err)
		}
	}
}

func (csw *ComputeServiceWrapper) deleteForwardingRule(project string, name string, region string) {
	_, err := csw.Compute.ForwardingRules.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) deleteGlobalForwardingRule(project string, name string) {
	_, err := csw.Compute.GlobalForwardingRules.Delete(project, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) deleteTargetPool(project string, name string, region string) {
	_, err := csw.Compute.TargetPools.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func (csw *ComputeServiceWrapper) lookupURLMaps(project string) ([]*compute.UrlMap, error) {
	call := csw.Compute.UrlMaps.List(project)
	var items []*compute.UrlMap
	f := func(page *compute.UrlMapList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (csw *ComputeServiceWrapper) lookupBackendServices(project string) ([]*compute.BackendService, error) {
	call := csw.Compute.BackendServices.List(project)
	var items []*compute.BackendService
	f := func(page *compute.BackendServiceList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (csw *ComputeServiceWrapper) lookupInstanceGroup(project string, zone string) ([]string, error) {
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

func (csw *ComputeServiceWrapper) lookupTargetPools(project string) ([]*compute.TargetPool, error) {
	call := csw.Compute.TargetPools.AggregatedList(project)
	var items []*compute.TargetPool
	f := func(page *compute.TargetPoolAggregatedList) error {
		for _, list := range page.Items {
			for _, element := range list.TargetPools {
				items = append(items, element)
			}
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
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

func (csw *ComputeServiceWrapper) lookupHTTPProxy(project string) ([]*compute.TargetHttpProxy, error) {
	call := csw.Compute.TargetHttpProxies.List(project)
	var items []*compute.TargetHttpProxy
	f := func(page *compute.TargetHttpProxyList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (csw *ComputeServiceWrapper) lookupGlobalForwardingRule(project string) ([]*compute.ForwardingRule, error) {
	call := csw.Compute.GlobalForwardingRules.List(project)
	var items []*compute.ForwardingRule
	f := func(page *compute.ForwardingRuleList) error {
		for _, list := range page.Items {
			items = append(items, list)
		}
		return nil
	}
	if err := call.Pages(csw.Context, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (csw *ComputeServiceWrapper) checkInstance(project string, zone string, name string) bool {
	_, err := csw.Compute.Instances.Get(project, zone, name).Do()
	if err != nil {
		return false
	}
	return true
}
