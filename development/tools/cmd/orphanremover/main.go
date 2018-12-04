// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"log"
	"os"
	"strings"
	"time"
)

var (
	project = flag.String("project", "", "Project ID")
	dryRun  = flag.Bool("dry-run", true, "Dry Run enabled")
)

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

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	var targetPool = []TargetPool{}
	var garbagePool = []TargetPool{}
	var instanceGroups = []InstanceGroup{}
	var backendServices = []BackendService{}
	var urlMaps = []URLMap{}
	var httpProxies = []HTTPProxy{}
	var globalForwardingRules = []GlobalForwardingRule{}

	ctx := context.Background()

	connenction, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := compute.New(connenction)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	log.Print("Creating mesh of network elements to delete. This takes some time\n")
	targetPool, err = lookupTargetPools(svc, *project)
	if err != nil {
		log.Fatalf("Could not list TargetPools: %v", err)
	}
	for _, target := range targetPool {
		markCount := 0
		for _, instance := range target.instances {
			markInstance(svc, *project, &instance)
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

	zones, err := lookupZones(svc, *project, "europe-*")
	if err != nil {
		log.Fatalf("Could not list Zones: %v", err)
	}
	for _, zone := range zones {
		igList, err := lookupInstanceGroup(svc, *project, zone)
		if err != nil {
			log.Fatalf("Could not list InstanceGroups: %v", err)
		}
		if len(igList) > 0 {
			for _, name := range igList {
				fields := strings.Split(name, "--")
				id := fields[len(fields)-1]
				instanceGroups = append(instanceGroups, InstanceGroup{name, id, zone})
			}
		}
	}
	urlMaps, err = lookupURLMaps(svc, *project)
	if err != nil {
		log.Fatalf("Could not list UrlMaps: %v", err)
	}
	backendServices, err = lookupBackendServices(svc, *project)
	if err != nil {
		log.Fatalf("Could not list BackendServices: %v", err)
	}
	httpProxies, err = lookupHTTPProxy(svc, *project)
	if err != nil {
		log.Fatalf("Could not list HTTPProxy: %v", err)
	}
	globalForwardingRules, err = lookupGlobalForwardingRule(svc, *project)
	if err != nil {
		log.Fatalf("Could not list GlobalForwardingRule: %v", err)
	}
	purge(svc, garbagePool, instanceGroups, backendServices, urlMaps, httpProxies, globalForwardingRules, *dryRun)
}

func purge(svc *compute.Service, targetPool []TargetPool, instanceGroups []InstanceGroup, backendServices []BackendService, urlMaps []URLMap, httpProxies []HTTPProxy, globalForwardingRules []GlobalForwardingRule, dryRun bool) {
	for _, target := range targetPool {
		fmt.Printf("-> Processing targetPool: %s\n", target.name)

		fmt.Printf("---> Delete ForwardingRules: %s in Region: %s\n", target.name, target.region)
		if !dryRun {
			deleteForwardingRule(svc, *project, target.name, target.region)
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete HealthCheck: %s\n", target.healthChecks)
		if !dryRun {
			deleteHealthChecks(svc, *project, target.healthChecks)
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete TargetPool: %s in Region: %s\n", target.name, target.region)
		if !dryRun {
			deleteTargetPool(svc, *project, target.name, target.region)
			time.Sleep(5 * time.Second)
		}
	}
	for _, group := range instanceGroups {
		fmt.Printf("-> Processing instanceGroup: %s\n", group.name)
		fmt.Printf("---> Delete ForwardingRules: %s\n", findGlobalForwardingRule(group.id, globalForwardingRules))
		if !dryRun {
			deleteGlobalForwardingRule(svc, *project, findGlobalForwardingRule(group.id, globalForwardingRules))
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete HTTPProxy: %s\n", findHTTPProxy(group.id, httpProxies))
		if !dryRun {
			deleteHTTPProxy(svc, *project, findHTTPProxy(group.id, httpProxies))
			time.Sleep(5 * time.Second)
		}
		fmt.Printf("---> Delete URLMap: %s\n", findURLMap(group.id, urlMaps))
		if !dryRun {
			deleteURLMap(svc, *project, findURLMap(group.id, urlMaps))
			time.Sleep(5 * time.Second)
		}
		services := findBackendServices(group.id, backendServices)
		for _, service := range services {
			fmt.Printf("---> Delete BackendService: %s\n", service)
			if !dryRun {
				deleteBackendService(svc, *project, service)
				time.Sleep(5 * time.Second)
			}
		}
		fmt.Printf("---> Delete InstanceGroup: %s in Zone: %s\n", group.name, group.zone)
		if !dryRun {
			deleteInstanceGroup(svc, *project, group.zone, group.name)
			time.Sleep(5 * time.Second)
		}
	}
}

func deleteHTTPProxy(svc *compute.Service, project string, httpProxy string) {
	_, err := svc.TargetHttpProxies.Delete(project, httpProxy).Do()
	if err != nil {
		log.Print(err)
	}
}

func deleteURLMap(svc *compute.Service, project string, urlMap string) {
	_, err := svc.UrlMaps.Delete(project, urlMap).Do()
	if err != nil {
		log.Print(err)
	}
}

func deleteBackendService(svc *compute.Service, project string, backendService string) {
	_, err := svc.BackendServices.Delete(project, backendService).Do()
	if err != nil {
		log.Print(err)
	}
}

func deleteInstanceGroup(svc *compute.Service, project string, zone string, instanceGroup string) {
	_, err := svc.InstanceGroups.Delete(project, zone, instanceGroup).Do()
	if err != nil {
		log.Print(err)
	}
}

func deleteHealthChecks(svc *compute.Service, project string, names []string) {
	for _, check := range names {
		_, err := svc.HttpHealthChecks.Delete(project, check).Do()
		if err != nil {
			log.Print(err)
		}
	}
}

func deleteForwardingRule(svc *compute.Service, project string, name string, region string) {
	_, err := svc.ForwardingRules.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func deleteGlobalForwardingRule(svc *compute.Service, project string, name string) {
	_, err := svc.GlobalForwardingRules.Delete(project, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func deleteTargetPool(svc *compute.Service, project string, name string, region string) {
	_, err := svc.TargetPools.Delete(project, region, name).Do()
	if err != nil {
		log.Print(err)
	}
}

func markInstance(svc *compute.Service, project string, instance *Instance) {
	_, err := svc.Instances.Get(project, instance.zone, instance.name).Do()
	if err != nil {
		instance.exists = false
	}
}

func lookupURLMaps(svc *compute.Service, project string) ([]URLMap, error) {
	call := svc.UrlMaps.List(project)
	var items []URLMap
	f := func(page *compute.UrlMapList) error {
		for _, list := range page.Items {
			fields := strings.Split(list.Name, "--")
			id := fields[len(fields)-1]
			items = append(items, URLMap{list.Name, id})
		}
		return nil
	}
	if err := call.Pages(context.Background(), f); err != nil {
		return nil, err
	}
	return items, nil
}

func lookupBackendServices(svc *compute.Service, project string) ([]BackendService, error) {
	call := svc.BackendServices.List(project)
	var items []BackendService
	f := func(page *compute.BackendServiceList) error {
		for _, list := range page.Items {
			fields := strings.Split(list.Name, "--")
			id := fields[len(fields)-1]
			items = append(items, BackendService{list.Name, id})
		}
		return nil
	}
	if err := call.Pages(context.Background(), f); err != nil {
		return nil, err
	}
	return items, nil
}

func lookupInstanceGroup(svc *compute.Service, project string, zone string) ([]string, error) {
	call := svc.InstanceGroups.List(project, zone)
	call = call.Filter("size: 0")
	var items []string
	f := func(page *compute.InstanceGroupList) error {
		for _, list := range page.Items {
			items = append(items, list.Name)
		}
		return nil
	}
	if err := call.Pages(context.Background(), f); err != nil {
		return nil, err
	}
	return items, nil
}

func lookupTargetPools(svc *compute.Service, project string) ([]TargetPool, error) {
	call := svc.TargetPools.AggregatedList(project)
	var items []TargetPool
	f := func(page *compute.TargetPoolAggregatedList) error {
		for _, list := range page.Items {
			for _, element := range list.TargetPools {
				region := strings.Split(element.Region, "/")
				var instances []Instance
				for _, inst := range element.Instances {
					fields := strings.Split(inst, "/")
					instances = append(instances, Instance{fields[len(fields)-1], fields[len(fields)-3], true})
				}
				var checks []string
				for _, check := range element.HealthChecks {
					chk := strings.Split(check, "/")
					checks = append(checks, chk[len(chk)-1])
				}
				item := TargetPool{element.Name, instances, len(instances), checks, region[len(region)-1], element.CreationTimestamp}

				items = append(items, item)
			}
		}
		return nil
	}
	if err := call.Pages(context.Background(), f); err != nil {
		return nil, err
	}
	return items, nil
}

func lookupZones(svc *compute.Service, project, pattern string) ([]string, error) {
	call := svc.Zones.List(project)
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

	if err := call.Pages(context.Background(), f); err != nil {
		return nil, err
	}
	return zones, nil
}

func lookupHTTPProxy(svc *compute.Service, project string) ([]HTTPProxy, error) {
	call := svc.TargetHttpProxies.List(project)
	var items []HTTPProxy
	f := func(page *compute.TargetHttpProxyList) error {
		for _, list := range page.Items {
			fields := strings.Split(list.Name, "--")
			id := fields[len(fields)-1]
			items = append(items, HTTPProxy{list.Name, id})
		}
		return nil
	}
	if err := call.Pages(context.Background(), f); err != nil {
		return nil, err
	}
	return items, nil
}

func lookupGlobalForwardingRule(svc *compute.Service, project string) ([]GlobalForwardingRule, error) {
	call := svc.GlobalForwardingRules.List(project)
	var items []GlobalForwardingRule
	f := func(page *compute.ForwardingRuleList) error {
		for _, list := range page.Items {
			fields := strings.Split(list.Name, "--")
			id := fields[len(fields)-1]
			items = append(items, GlobalForwardingRule{list.Name, id})
		}
		return nil
	}
	if err := call.Pages(context.Background(), f); err != nil {
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
