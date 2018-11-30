// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"log"
	"os"
	"strings"
	"time"
)

var (
	project        = flag.String("project", "", "Project ID")
	dryRun         = flag.Bool("dry-run", true, "Dry Run enabled")
	targetPool     = []TargetPool{}
	garbagePool    = []TargetPool{}
	instanceGroups = []InstanceGroup{}
)

//TargetPool ???
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
	zone string
}

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	context := context.Background()
	connenction, err := google.DefaultClient(context, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := compute.New(connenction)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	// targetPool, _ := lookupTargetPools(svc, *project)
	// for _, target := range targetPool {
	// 	markCount := 0
	// 	for _, instance := range target.instances {
	// 		markInstance(svc, *project, &instance)
	// 		if !instance.exists {
	// 			markCount++
	// 		}
	// 	}
	// 	if markCount == target.instanceCount {
	// 		garbagePool = append(garbagePool, target)
	// 	}
	// }
	// fmt.Printf("All items: %d\n", len(targetPool))
	// fmt.Printf("Garbage items: %d\n", len(garbagePool))

	zones, _ := lookupZones(svc, *project, "europe-*")
	for _, zone := range zones {
		igList, _ := lookupInstanceGroup(svc, *project, zone)
		if len(igList) > 0 {
			for _, ig := range igList {
				// fmt.Printf("IG: %s, ZONE: %s\n", ig, zone)
				instanceGroups = append(instanceGroups, InstanceGroup{ig, zone})
			}

		}
	}
	fmt.Printf("%s", instanceGroups)
	deleteInstanceGroup(svc, *project, instanceGroups[0].zone, instanceGroups[0].name)
	// log.Print("---> Looking up BackendServices!")
	// lookupBackendServices(svc, *project, garbagePool[0].healthChecks)

	// log.Print("---> Looking up InstanceGroups!\n Item: %s", garbagePool[0])
	// lookupInstanceGroup(svc, *project, garbagePool[0].instances[0].zone)

	if !*dryRun {
		purge(svc, garbagePool)
	}
}

func purge(svc *compute.Service, pool []TargetPool) {
	for _, target := range pool {
		log.Print("Processing item: %s\n", target.name)

		log.Print("Delete ForwardingRules: %s\n", target.name)
		deleteForwardingRule(svc, *project, target.name, target.region)
		time.Sleep(5 * time.Second)

		log.Print("Delete HealthCheck : %s\n", target.healthChecks)
		deleteHealthChecks(svc, *project, target.healthChecks)
		time.Sleep(5 * time.Second)

		log.Print("Delete TargetPool : %s\n", target.name)
		deleteTargetPool(svc, *project, target.name, target.region)
		time.Sleep(5 * time.Second)
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

func lookupBackendServices(svc *compute.Service, project string, checks []string) ([]string, error) {
	call := svc.BackendServices.List(project)
	var items []string
	f := func(page *compute.BackendServiceList) error {
		for _, list := range page.Items {
			js, _ := json.Marshal(list)
			fmt.Printf("%s\n\n", js)
		}
		return nil
	}
	if err := call.Pages(oauth2.NoContext, f); err != nil {
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
	if err := call.Pages(oauth2.NoContext, f); err != nil {
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
					singleInstance := Instance{fields[len(fields)-1], fields[len(fields)-3], true}
					instances = append(instances, singleInstance)
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
	if err := call.Pages(oauth2.NoContext, f); err != nil {
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

	if err := call.Pages(oauth2.NoContext, f); err != nil {
		return nil, err
	}
	return zones, nil
}
