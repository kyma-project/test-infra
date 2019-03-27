// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/kyma-project/test-infra/development/tools/pkg/clusterscollector"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	container "google.golang.org/api/container/v1"
)

const defaultClusterNameRegexp = "^gkeint[-](pr|commit|rel)[-].*"

var (
	project           = flag.String("project", "", "Project ID [Required]")
	dryRun            = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	strategy          = flag.String("strategy", "default", "Change cluster filter strategy. Current options are 'default' and 'time'.")
	ageInHours        = flag.Int("ageInHours", 3, "Cluster age in hours. Clusters older than: now()-ageInHours are subject to removal.")
	clusterNameRegexp = flag.String("clusterNameRegexp", defaultClusterNameRegexp, "Cluster name regexp pattern. Matching clusters are subject to removal.")
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

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ageInHours: %d, clusterNameRegexp: \"%s\"", *project, *dryRun, *ageInHours, *clusterNameRegexp)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	containerSvc, err := container.New(connection)
	if err != nil {
		log.Fatalf("Could not initialize container API client: %v", err)
	}

	clusterNameRx := regexp.MustCompile(*clusterNameRegexp)

	clustersService := containerSvc.Projects.Locations.Clusters

	clusterAPI := &clusterscollector.ClusterAPIWrapper{Context: ctx, Service: clustersService}

	var clusterFilter clusterscollector.ClusterRemovalPredicate

	if *strategy == "time" {
		clusterFilter = clusterscollector.TimeBasedClusterRemovalPredicate()
		log.Infof("Using time based filter strategy. Clusters will be filtered based on TTL, volatility, created-at timestamp and status\n")
	} else {
		log.Infof("Using default filter strategy. Clusters will be filtered based on cluster name, age in hours (passed), volatility and status\n")
		clusterFilter = clusterscollector.DefaultClusterRemovalPredicate(clusterNameRx, uint(*ageInHours))
	}
	gc := clusterscollector.NewClustersGarbageCollector(clusterAPI, clusterFilter)

	allSucceeded, err := gc.Run(*project, !(*dryRun))

	if err != nil {
		log.Fatalf("Cluster collector error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}
