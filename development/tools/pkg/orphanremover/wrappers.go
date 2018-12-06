package orphanremover

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"
)

// ComputeAPIWrapper ???
type ComputeAPIWrapper struct {
	Ctx context.Context
	Svc *compute.Service
}

// TargetPool ???
type TargetPool struct {
	name          string
	instances     []Instance
	instanceCount int
	healthChecks  []string
	region        string
	timestamp     string
}

// Instance ???
type Instance struct {
	name   string
	zone   string
	exists bool
}

// InstanceGroup ???
type InstanceGroup struct {
	name string
	id   string
	zone string
}

// BackendService ???
type BackendService struct {
	name string
	id   string
}

// URLMap ???
type URLMap struct {
	name string
	id   string
}

// HTTPProxy ???
type HTTPProxy struct {
	name string
	id   string
}

// GlobalForwardingRule ???
type GlobalForwardingRule struct {
	name string
	id   string
}

// Collect ???
func (computeAPIWrapper *ComputeAPIWrapper) Collect(dryRun bool, project string) {
	var garbagePool = []TargetPool{}
	var instanceGroups = []InstanceGroup{}

	log.Print("Creating mesh of network elements to delete. This takes some time\n")
	targetPool, err := computeAPIWrapper.lookupTargetPools(project)
	if err != nil {
		log.Fatalf("Could not list TargetPools: %v", err)
	}
	for _, target := range targetPool {
		markCount := 0
		for _, instance := range target.instances {
			computeAPIWrapper.markInstance(project, &instance)
			if !instance.exists {
				markCount++
			}
		}
		if markCount == target.instanceCount {
			garbagePool = append(garbagePool, target)
		}
	}

	fmt.Printf("All items: %d\n", len(targetPool))
	fmt.Printf("Garbage items: %d\n", len(garbagePool))

	zones, err := computeAPIWrapper.lookupZones(project, "europe-*")
	if err != nil {
		log.Fatalf("Could not list Zones: %v", err)
	}
	for _, zone := range zones {
		igList, err := computeAPIWrapper.lookupInstanceGroup(project, zone)
		if err != nil {
			log.Fatalf("Could not list InstanceGroups: %v", err)
		}
		if len(igList) > 0 {
			for _, name := range igList {
				instanceGroups = append(instanceGroups, InstanceGroup{name, spliter(name, "--", 1), zone})
			}
		}
	}
	urlMaps, err := computeAPIWrapper.lookupURLMaps(project)
	if err != nil {
		log.Fatalf("Could not list UrlMaps: %v", err)
	}
	backendServices, err := computeAPIWrapper.lookupBackendServices(project)
	if err != nil {
		log.Fatalf("Could not list BackendServices: %v", err)
	}
	httpProxies, err := computeAPIWrapper.lookupHTTPProxy(project)
	if err != nil {
		log.Fatalf("Could not list HTTPProxy: %v", err)
	}
	globalForwardingRules, err := computeAPIWrapper.lookupGlobalForwardingRule(project)
	if err != nil {
		log.Fatalf("Could not list GlobalForwardingRule: %v", err)
	}
	computeAPIWrapper.purge(garbagePool, instanceGroups, backendServices, urlMaps, httpProxies, globalForwardingRules, project, dryRun)
}

func (computeAPIWrapper *ComputeAPIWrapper) purge(targetPool []TargetPool, instanceGroups []InstanceGroup, backendServices []BackendService, urlMaps []URLMap, httpProxies []HTTPProxy, globalForwardingRules []GlobalForwardingRule, project string, dryRun bool) {
	for _, target := range targetPool {
		fmt.Printf("-> Processing targetPool: %s\n", target.name)

		fmt.Printf("---> Delete ForwardingRules: %s in Region: %s\n", target.name, target.region)
		if !dryRun {
			computeAPIWrapper.deleteForwardingRule(project, target.name, target.region)
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete HealthCheck: %s\n", target.healthChecks)
		if !dryRun {
			computeAPIWrapper.deleteHealthChecks(project, target.healthChecks)
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete TargetPool: %s in Region: %s\n", target.name, target.region)
		if !dryRun {
			computeAPIWrapper.deleteTargetPool(project, target.name, target.region)
			time.Sleep(5 * time.Second)
		}
	}

	for _, group := range instanceGroups {
		fmt.Printf("-> Processing instanceGroup: %s\n", group.name)
		fmt.Printf("---> Delete ForwardingRules: %s\n", findGlobalForwardingRule(group.id, globalForwardingRules))
		if !dryRun {
			computeAPIWrapper.deleteGlobalForwardingRule(project, findGlobalForwardingRule(group.id, globalForwardingRules))
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete HTTPProxy: %s\n", findHTTPProxy(group.id, httpProxies))
		if !dryRun {
			computeAPIWrapper.deleteHTTPProxy(project, findHTTPProxy(group.id, httpProxies))
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete URLMap: %s\n", findURLMap(group.id, urlMaps))
		if !dryRun {
			computeAPIWrapper.deleteURLMap(project, findURLMap(group.id, urlMaps))
			time.Sleep(5 * time.Second)
		}
		services := findBackendServices(group.id, backendServices)
		for _, service := range services {
			fmt.Printf("---> Delete BackendService: %s\n", service)
			if !dryRun {
				computeAPIWrapper.deleteBackendService(project, service)
				time.Sleep(5 * time.Second)
			}
		}
		fmt.Printf("---> Delete InstanceGroup: %s in Zone: %s\n", group.name, group.zone)
		if !dryRun {
			computeAPIWrapper.deleteInstanceGroup(project, group.zone, group.name)
			time.Sleep(5 * time.Second)
		}
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) markInstance(project string, instance *Instance) {
	_, err := computeAPIWrapper.Svc.Instances.Get(project, instance.zone, instance.name).Do()
	if err != nil {
		instance.exists = false
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteHTTPProxy(project string, httpProxy string) {
	_, err := computeAPIWrapper.Svc.TargetHttpProxies.Delete(project, httpProxy).Do()
	if err != nil {
		log.Print(err)
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteURLMap(project string, urlMap string) {
	_, err := computeAPIWrapper.Svc.UrlMaps.Delete(project, urlMap).Do()
	if err != nil {
		log.Print(err)
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteBackendService(project string, backendService string) {
	_, err := computeAPIWrapper.Svc.BackendServices.Delete(project, backendService).Do()
	if err != nil {
		log.Print(err)
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteInstanceGroup(project string, zone string, instanceGroup string) {
	_, err := computeAPIWrapper.Svc.InstanceGroups.Delete(project, zone, instanceGroup).Do()
	if err != nil {
		log.Print(err)
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteHealthChecks(project string, names []string) {
	for _, check := range names {
		_, err := computeAPIWrapper.Svc.HttpHealthChecks.Delete(project, check).Do()
		if err != nil {
			log.Print(err)
		}
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteForwardingRule(project string, name string, region string) {
	_, err := computeAPIWrapper.Svc.ForwardingRules.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteGlobalForwardingRule(project string, name string) {
	_, err := computeAPIWrapper.Svc.GlobalForwardingRules.Delete(project, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func (computeAPIWrapper *ComputeAPIWrapper) deleteTargetPool(project string, name string, region string) {
	_, err := computeAPIWrapper.Svc.TargetPools.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func spliter(name string, delimiter string, position int) string {
	fields := strings.Split(name, delimiter)
	return fields[len(fields)-position]
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupURLMaps(project string) ([]URLMap, error) {
	call := computeAPIWrapper.Svc.UrlMaps.List(project)
	var items []URLMap
	f := func(page *compute.UrlMapList) error {
		for _, list := range page.Items {
			items = append(items, URLMap{list.Name, spliter(list.Name, "--", 1)})
		}
		return nil
	}
	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupBackendServices(project string) ([]BackendService, error) {
	call := computeAPIWrapper.Svc.BackendServices.List(project)
	var items []BackendService
	f := func(page *compute.BackendServiceList) error {
		for _, list := range page.Items {
			items = append(items, BackendService{list.Name, spliter(list.Name, "--", 1)})
		}
		return nil
	}
	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupInstanceGroup(project string, zone string) ([]string, error) {
	call := computeAPIWrapper.Svc.InstanceGroups.List(project, zone)
	call = call.Filter("size: 0")
	var items []string
	f := func(page *compute.InstanceGroupList) error {
		for _, list := range page.Items {
			items = append(items, list.Name)
		}
		return nil
	}
	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupTargetPools(project string) ([]TargetPool, error) {
	call := computeAPIWrapper.Svc.TargetPools.AggregatedList(project)
	var items []TargetPool
	f := func(page *compute.TargetPoolAggregatedList) error {
		for _, list := range page.Items {
			for _, element := range list.TargetPools {
				var instances []Instance
				for _, inst := range element.Instances {
					instances = append(instances, Instance{spliter(inst, "/", 1), spliter(inst, "/", 3), true})
				}
				var checks []string
				for _, check := range element.HealthChecks {
					checks = append(checks, spliter(check, "/", 1))
				}
				item := TargetPool{element.Name, instances, len(instances), checks, spliter(element.Region, "/", 1), element.CreationTimestamp}
				items = append(items, item)
			}
		}
		return nil
	}
	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupZones(project, pattern string) ([]string, error) {
	call := computeAPIWrapper.Svc.Zones.List(project)
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

	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return zones, nil
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupHTTPProxy(project string) ([]HTTPProxy, error) {
	call := computeAPIWrapper.Svc.TargetHttpProxies.List(project)
	var items []HTTPProxy
	f := func(page *compute.TargetHttpProxyList) error {
		for _, list := range page.Items {
			items = append(items, HTTPProxy{list.Name, spliter(list.Name, "--", 1)})
		}
		return nil
	}
	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return items, nil
}

func (computeAPIWrapper *ComputeAPIWrapper) lookupGlobalForwardingRule(project string) ([]GlobalForwardingRule, error) {
	call := computeAPIWrapper.Svc.GlobalForwardingRules.List(project)
	var items []GlobalForwardingRule
	f := func(page *compute.ForwardingRuleList) error {
		for _, list := range page.Items {
			items = append(items, GlobalForwardingRule{list.Name, spliter(list.Name, "--", 1)})
		}
		return nil
	}
	if err := call.Pages(computeAPIWrapper.Ctx, f); err != nil {
		return nil, err
	}
	return items, nil
}

func findBackendServices(id string, backendServices []BackendService) []string {
	var items []string
	for _, service := range backendServices {
		if service.id == id {
			items = append(items, service.name)
		}
	}
	return items
}

func findURLMap(id string, urlMaps []URLMap) string {
	for _, maps := range urlMaps {
		if maps.id == id {
			return maps.name
		}
	}
	return "NOT_FOUND"
}

func findHTTPProxy(id string, httpProxy []HTTPProxy) string {
	for _, proxy := range httpProxy {
		if proxy.id == id {
			return proxy.name
		}
	}
	return "NOT_FOUND"
}

func findGlobalForwardingRule(id string, forwadingRules []GlobalForwardingRule) string {
	for _, rule := range forwadingRules {
		if rule.id == id {
			return rule.name
		}
	}
	return "NOT_FOUND"
}
