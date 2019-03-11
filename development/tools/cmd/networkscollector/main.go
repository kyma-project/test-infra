package main

import (
	"context"
	"flag"
	"fmt"

	"os"
	"regexp"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/networkscollector"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	container "google.golang.org/api/container/v1"
)

const defaultNetworkNameRegexp = "^net-gkeint[-](pr|commit|rel)[-].*"

var (
	project           = flag.String("project", "", "Project ID [Required]")
	dryRun            = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	ageInHours        = flag.Int("ageInHours", 3, "Network age in hours. Networks older than: now()-ageInHours are subject to removal.")
	networkNameRegexp = flag.String("networkNameRegexp", defaultNetworkNameRegexp, "Network name regexp pattern. Matching networks are subject to removal.")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *ageInHours < 1 {
		fmt.Fprintln(os.Stderr, "invalid ageInHours value, must be greater than zero")
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ageInHours: %d", *project, *dryRun, *ageInHours)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, compute.CloudPlatformScope, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	computeSvc, err := compute.New(connection)
	if err != nil {
		log.Fatalf("Could not initialize compute API client: %v", err)
	}

	networkNameRx := regexp.MustCompile(*networkNameRegexp)

	networksSvc := computeSvc.Networks

	networkAPI := &networkscollector.NetworkAPIWrapper{Context: ctx, Service: networksSvc}

	networkFilter := networkscollector.DefaultNetworkRemovalPredicate(networkNameRx, uint(*ageInHours))
	gc := networkscollector.NewNetworksGarbageCollector(networkAPI, networkFilter)

	allSucceeded, err := gc.Run(*project, !(*dryRun))

	if err != nil {
		log.Fatalf("Network collector error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}
