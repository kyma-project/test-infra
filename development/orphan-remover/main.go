// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"log"
	"os"
	"strings"
)

var (
	project     = flag.String("project", "", "Project ID")
	targetPool  = []TargetPool{}
	garbagePool = []TargetPool{}
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

	targetPool, _ := lookupTargetPools(svc, *project)
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
	fmt.Printf("Processing item: %s\n", garbagePool[0].name)
	fmt.Printf("HealthChecks item: %s\n", garbagePool[0].healthChecks)
	// deleteForwardingRule(svc, *project, garbagePool[0].name, garbagePool[0].region)
	// deleteTargetPool(svc, *project, garbagePool[0].name, garbagePool[0].region)
	deleteHealthChecks(svc, *project, garbagePool[0].healthChecks)
}

func deleteHealthChecks(svc *compute.Service, project string, names []string) {
	for _, check := range names {
		_, err := svc.HealthChecks.Delete(project, name).Do()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func deleteForwardingRule(svc *compute.Service, project string, name string, region string) {
	_, err := svc.ForwardingRules.Delete(project, region, name).Do()
	if err != nil {
		log.Fatal(err)
	}
}

func deleteTargetPool(svc *compute.Service, project string, name string, region string) {
	_, err := svc.TargetPools.Delete(project, region, name).Do()
	if err != nil {
		log.Fatal(err)
	}
}

func markInstance(svc *compute.Service, project string, instance *Instance) {
	_, err := svc.Instances.Get(project, instance.zone, instance.name).Do()
	if err != nil {
		instance.exists = false
	}
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
