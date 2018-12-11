package orphanremover

import (
	"fmt"
	"log"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"
)

const sleepFactor = 2

//ComputeAPI interface logic for Google cloud API
type ComputeAPI interface {
	deleteHTTPProxy(project string, httpProxy string)
	deleteURLMap(project string, urlMap string)
	deleteBackendService(project string, backendService string)
	deleteInstanceGroup(project string, zone string, instanceGroup string)
	deleteHealthChecks(project string, names []string)
	deleteForwardingRule(project string, name string, region string)
	deleteGlobalForwardingRule(project string, name string)
	deleteTargetPool(project string, name string, region string)
	lookupURLMaps(project string) ([]*compute.UrlMap, error)
	lookupBackendServices(project string) ([]*compute.BackendService, error)
	lookupInstanceGroup(project string, zone string) ([]string, error)
	lookupTargetPools(project string) ([]*compute.TargetPool, error)
	lookupZones(project, pattern string) ([]string, error)
	lookupHTTPProxy(project string) ([]*compute.TargetHttpProxy, error)
	lookupGlobalForwardingRule(project string) ([]*compute.ForwardingRule, error)
	checkInstance(project string, zone string, name string) bool
}

//Remover Element holding the removal logic
type Remover struct {
	computeAPI ComputeAPI
}

type targetPool struct {
	name          string
	instances     []instance
	instanceCount int
	healthChecks  []string
	region        string
	timestamp     string
}

type instance struct {
	name   string
	zone   string
	exists bool
}

type instanceGroup struct {
	name string
	id   string
	zone string
}

type backendService struct {
	name string
	id   string
}

type urlMap struct {
	name string
	id   string
}

type httpProxy struct {
	name string
	id   string
}

type globalForwardingRule struct {
	name string
	id   string
}

//NewRemover Returns a new remover object
func NewRemover(computeAPI ComputeAPI) *Remover {
	return &Remover{computeAPI}
}

func spliter(name string, delimiter string, position int) string {
	fields := strings.Split(name, delimiter)
	return fields[len(fields)-position]
}

//Run the main find&destroy function
func (remover *Remover) Run(dryRun bool, project string) {
	var garbagePool = []targetPool{}
	var instanceGroups = []instanceGroup{}

	log.Print("Creating mesh of network elements to delete. This takes some time\n")
	rawTargetPool, err := remover.computeAPI.lookupTargetPools(project)
	if err != nil {
		log.Fatalf("Could not list TargetPools: %v", err)
	}
	targetPool := extractTargetPool(rawTargetPool)
	for _, target := range targetPool {
		markCount := 0
		for _, instance := range target.instances {
			instance.exists = remover.computeAPI.checkInstance(project, instance.zone, instance.name)
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

	zones, err := remover.computeAPI.lookupZones(project, "europe-*")
	if err != nil {
		log.Fatalf("Could not list Zones: %v", err)
	}
	for _, zone := range zones {
		igList, err := remover.computeAPI.lookupInstanceGroup(project, zone)
		if err != nil {
			log.Fatalf("Could not list InstanceGroups: %v", err)
		}
		if len(igList) > 0 {
			for _, name := range igList {
				instanceGroups = append(instanceGroups, instanceGroup{name, spliter(name, "--", 1), zone})
			}
		}
	}
	rawURLMaps, err := remover.computeAPI.lookupURLMaps(project)
	if err != nil {
		log.Fatalf("Could not list UrlMaps: %v", err)
	}
	urlMaps := extractURLMaps(rawURLMaps)

	rawBackendServices, err := remover.computeAPI.lookupBackendServices(project)
	if err != nil {
		log.Fatalf("Could not list BackendServices: %v", err)
	}
	backendServices := extractBackendServices(rawBackendServices)

	rawHTTPProxies, err := remover.computeAPI.lookupHTTPProxy(project)
	if err != nil {
		log.Fatalf("Could not list HTTPProxy: %v", err)
	}
	httpProxies := extractHTTPProxies(rawHTTPProxies)

	rawGlobalForwardingRules, err := remover.computeAPI.lookupGlobalForwardingRule(project)
	if err != nil {
		log.Fatalf("Could not list GlobalForwardingRule: %v", err)
	}
	globalForwardingRules := extractForwardingRules(rawGlobalForwardingRules)

	remover.purge(garbagePool, instanceGroups, backendServices, urlMaps, httpProxies, globalForwardingRules, project, dryRun)
}

func (remover *Remover) purge(targetPool []targetPool, instanceGroups []instanceGroup, backendServices []backendService, urlMaps []urlMap, httpProxies []httpProxy, globalForwardingRules []globalForwardingRule, project string, dryRun bool) {
	for _, target := range targetPool {
		fmt.Printf("-> Processing targetPool: %s\n", target.name)

		fmt.Printf("---> Delete ForwardingRules: %s in Region: %s\n", target.name, target.region)
		if !dryRun {
			remover.computeAPI.deleteForwardingRule(project, target.name, target.region)
			time.Sleep(sleepFactor * time.Second)
		}
		fmt.Printf("---> Delete HealthCheck: %s\n", target.healthChecks)
		if !dryRun {
			remover.computeAPI.deleteHealthChecks(project, target.healthChecks)
			time.Sleep(sleepFactor * time.Second)
		}
		fmt.Printf("---> Delete TargetPool: %s in Region: %s\n", target.name, target.region)
		if !dryRun {
			remover.computeAPI.deleteTargetPool(project, target.name, target.region)
			time.Sleep(sleepFactor * time.Second)
		}
	}

	for _, group := range instanceGroups {
		fmt.Printf("-> Processing instanceGroup: %s\n", group.name)
		fmt.Printf("---> Delete ForwardingRules: %s\n", findGlobalForwardingRule(group.id, globalForwardingRules))
		if !dryRun {
			remover.computeAPI.deleteGlobalForwardingRule(project, findGlobalForwardingRule(group.id, globalForwardingRules))
			time.Sleep(sleepFactor * time.Second)
		}
		fmt.Printf("---> Delete HTTPProxy: %s\n", findHTTPProxy(group.id, httpProxies))
		if !dryRun {
			remover.computeAPI.deleteHTTPProxy(project, findHTTPProxy(group.id, httpProxies))
			time.Sleep(sleepFactor * time.Second)
		}
		fmt.Printf("---> Delete URLMap: %s\n", findURLMap(group.id, urlMaps))
		if !dryRun {
			remover.computeAPI.deleteURLMap(project, findURLMap(group.id, urlMaps))
			time.Sleep(sleepFactor * time.Second)
		}
		services := findBackendServices(group.id, backendServices)
		for _, service := range services {
			fmt.Printf("---> Delete BackendService: %s\n", service)
			if !dryRun {
				remover.computeAPI.deleteBackendService(project, service)
				time.Sleep(sleepFactor * time.Second)
			}
		}
		fmt.Printf("---> Delete InstanceGroup: %s in Zone: %s\n", group.name, group.zone)
		if !dryRun {
			remover.computeAPI.deleteInstanceGroup(project, group.zone, group.name)
			time.Sleep(sleepFactor * time.Second)
		}
	}
}

func extractForwardingRules(rules []*compute.ForwardingRule) []globalForwardingRule {
	var items []globalForwardingRule
	for _, rule := range rules {
		items = append(items, globalForwardingRule{rule.Name, spliter(rule.Name, "--", 1)})
	}
	return items
}

func extractHTTPProxies(proxies []*compute.TargetHttpProxy) []httpProxy {
	var items []httpProxy
	for _, proxy := range proxies {
		items = append(items, httpProxy{proxy.Name, spliter(proxy.Name, "--", 1)})
	}
	return items
}

func extractBackendServices(services []*compute.BackendService) []backendService {
	var items []backendService
	for _, bs := range services {
		items = append(items, backendService{bs.Name, spliter(bs.Name, "--", 1)})
	}
	return items
}

func extractTargetPool(pool []*compute.TargetPool) []targetPool {
	var items []targetPool
	for _, target := range pool {
		var instances []instance
		for _, inst := range target.Instances {
			instances = append(instances, instance{spliter(inst, "/", 1), spliter(inst, "/", 3), true})
		}
		var checks []string
		for _, check := range target.HealthChecks {
			checks = append(checks, spliter(check, "/", 1))
		}
		item := targetPool{target.Name, instances, len(instances), checks, spliter(target.Region, "/", 1), target.CreationTimestamp}
		items = append(items, item)
	}
	return items
}

func extractURLMaps(maps []*compute.UrlMap) []urlMap {
	var items []urlMap
	for _, url := range maps {
		items = append(items, urlMap{url.Name, spliter(url.Name, "--", 1)})
	}
	return items
}

func findBackendServices(id string, backendServices []backendService) []string {
	var items []string
	for _, service := range backendServices {
		if service.id == id {
			items = append(items, service.name)
		}
	}
	return items
}

func findURLMap(id string, urlMaps []urlMap) string {
	for _, maps := range urlMaps {
		if maps.id == id {
			return maps.name
		}
	}
	return "NOT_FOUND"
}

func findHTTPProxy(id string, httpProxy []httpProxy) string {
	for _, proxy := range httpProxy {
		if proxy.id == id {
			return proxy.name
		}
	}
	return "NOT_FOUND"
}

func findGlobalForwardingRule(id string, forwadingRules []globalForwardingRule) string {
	for _, rule := range forwadingRules {
		if rule.id == id {
			return rule.name
		}
	}
	return "NOT_FOUND"
}
