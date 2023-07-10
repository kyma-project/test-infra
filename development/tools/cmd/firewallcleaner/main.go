// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/tools/pkg/firewallcleaner"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

var (
	project = flag.String("project", "", "Project ID [Required]")
	dryRun  = flag.Bool("dryRun", true, "Dry Run enabled")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	log.Printf("Running with arguments: project: \"%s\", dryRun: \"%t\"", *project, *dryRun)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	computeSvc, err := compute.NewService(ctx, option.WithHTTPClient(connection))
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	containerSvc, err := container.NewService(ctx, option.WithHTTPClient(connection))
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	computeServiceWrapper := &firewallcleaner.ComputeServiceWrapper{Context: ctx, Compute: computeSvc, Container: containerSvc}
	cleaner := firewallcleaner.NewCleaner(computeServiceWrapper)
	cleanerErr := cleaner.Run(*dryRun, *project)
	if cleanerErr != nil {
		log.Println(cleanerErr)
	}
}
