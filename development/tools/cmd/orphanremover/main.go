// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/orphanremover"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

var (
	project = flag.String("project", "", "Project ID")
	dryRun  = flag.Bool("dry-run", true, "Dry Run enabled")
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
	connenction, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := compute.New(connenction)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	computeAPIWrapper := &orphanremover.ComputeAPIWrapper{Ctx: ctx, Svc: svc}
	computeAPIWrapper.Collect(*dryRun, *project)
}
