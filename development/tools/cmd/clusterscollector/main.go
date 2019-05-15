// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/clusterscollector"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
)

var (
	project             = flag.String("project", "", "Project ID [Required]")
	dryRun              = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	ageInHours          = flag.Int("ageInHours", 3, "Cluster age in hours. Clusters older than: now()-ageInHours are subject to removal.")
	whitelistedClusters = flag.String("whitelisted-clusters", "", "Comma separated list of the whitelisted clusters that cannot be removed by cluster collector")
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

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ageInHours: %d, whitelisted clusters: %v", *project, *dryRun, *ageInHours, *whitelistedClusters)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	containerSvc, err := container.New(connection)
	if err != nil {
		log.Fatalf("Could not initialize container API client: %v", err)
	}

	clustersService := containerSvc.Projects.Locations.Clusters

	clusterAPI := &clusterscollector.ClusterAPIWrapper{Context: ctx, Service: clustersService}

	var clusterFilter clusterscollector.ClusterRemovalPredicate

	whClustersMap := map[string]struct{}{}
	for _, cl := range strings.Split(*whitelistedClusters, ",") {
		whClustersMap[cl] = struct{}{}
	}

	clusterFilter = clusterscollector.TimeBasedClusterRemovalPredicate(whClustersMap)
	log.Infof("Using time based filter strategy. Clusters will be filtered based on TTL, volatility, created-at timestamp and status\n")

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
