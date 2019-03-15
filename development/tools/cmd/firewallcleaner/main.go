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

	"github.com/kyma-project/test-infra/development/tools/pkg/firewallcleaner"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

var (
	project           = flag.String("project", "", "Project ID [Required]")
	dryRun            = flag.Bool("dryRun", true, "Dry Run enabled")
	githubRepoOwner   = flag.String("githubRepoOwner", "", "Github repository owner [Required]")
	githubRepoName    = flag.String("githubRepoName", "", "Github repository name [Required]")
	githubAccessToken = flag.String("githubAccessToken", "", "Github access token [Required]")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubRepoOwner == "" {
		fmt.Fprintln(os.Stderr, "missing -githubRepoOwner flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubRepoName == "" {
		fmt.Fprintln(os.Stderr, "missing -githubRepoName flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubAccessToken == "" {
		fmt.Fprintln(os.Stderr, "missing -githubAccessToken flag")
		flag.Usage()
		os.Exit(2)
	}

	log.Printf("Running with arguments: project: \"%s\", dryRun: \"%t\"", *project, *dryRun)
	ctx := context.Background()

	ga := firewallcleaner.NewGithubAPI(ctx, *githubAccessToken, *githubRepoOwner, *githubRepoName)

	connenction, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := compute.New(connenction)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	computeServiceWrapper := &firewallcleaner.ComputeServiceWrapper{Context: ctx, Compute: svc}
	cleaner := firewallcleaner.NewCleaner(computeServiceWrapper, ga)
	cleaner.Run(*dryRun, *project)
}
