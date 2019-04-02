// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/iprelease"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

const defaultClusterNameRegexp = "^gkeint[-](pr|commit|rel)[-].*"

var (
	project = flag.String("project", "", "Project ID [Required]")
	ipName  = flag.String("ipname", "", "IP resource name [Required]")
	dryRun  = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *ipName == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ipname: \"%s\"", *project, *dryRun, *ipName)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	computeSvc, err := compute.New(connection)
	if err != nil {
		log.Fatalf("Could not initialize compute API client: %v", err)
	}

	computeAPI := &iprelease.ComputeAPIWrapper{Context: ctx, Service: computeSvc}

	allSucceeded, err := computeAPI.Run(*project, ipName, !(*dryRun))

	if err != nil {
		log.Fatalf("Cluster collector error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}
