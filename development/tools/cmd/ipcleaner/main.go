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
	"github.com/kyma-project/test-infra/development/tools/pkg/ipcleaner"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

var (
	project     = flag.String("project", "", "Project ID [Required]")
	ipName      = flag.String("ipname", "", "IP resource name [Required]")
	region      = flag.String("region", "", "Region name [Required]")
	maxAttempts = flag.Uint("attempts", 3, "Maximal number of attempts until scripts stops trying to delete IP (default: 3)")
	backoff     = flag.Uint("backoff", 5, "Initial backoff in seconds for the first retry, will increase after this (default: 5)")
	dryRun      = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *ipName == "" {
		fmt.Fprintln(os.Stderr, "missing -ipname flag")
		flag.Usage()
		os.Exit(2)
	}

	if *region == "" {
		fmt.Fprintln(os.Stderr, "missing -region flag")
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

	computeAPI := &ipcleaner.ComputeAPIWrapper{Context: ctx, Service: computeSvc}

	ipr := ipcleaner.New(computeAPI, *maxAttempts, *backoff, !(*dryRun))

	success, err := ipr.Run(*project, *region, *ipName)

	if err != nil {
		log.Fatalf("IP Cleaner error: %v", err)
	}

	if !success {
		log.Warn("Operation failed.")
	}

	common.Shout("Finished")
}
