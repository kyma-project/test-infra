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

	"github.com/kyma-project/test-infra/pkg/tools/common"
	"github.com/kyma-project/test-infra/pkg/tools/gcrcleaner"

	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
	log "github.com/sirupsen/logrus"
)

const defaultGcrNameIgnoreRegex = ""

var (
	repository         = flag.String("repository", "", "Name of GCR repository [Required]")
	dryRun             = flag.Bool("dry-run", true, "Dry Run enabled, nothing is deleted")
	ageInHours         = flag.Int("age-in-hours", 24, "Address age in hours. images older than: now()-ageInHours are considered for removal.")
	gcrNameIgnoreRegex = flag.String("gcr-exclude-name-regex", defaultGcrNameIgnoreRegex, "Ignored GCR name regex. Matching paths are not considered for removal.")
)

func main() {
	flag.Parse()

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatalf("Requires the environment variable GOOGLE_APPLICATION_CREDENTIALS to be set to a GCP service account file.")
	}

	if *repository == "" {
		fmt.Fprintln(os.Stderr, "missing -repository flag")
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: repository: \"%s\", dryRun: %t, ageInHours: %d, gcrNameIgnoreRegex: \"%s\"", *repository, *dryRun, *ageInHours, *gcrNameIgnoreRegex)

	ctx := context.Background()

	auth := gcrgoogle.NewJSONKeyAuthenticator(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	regexRepo := regexp.MustCompile(*gcrNameIgnoreRegex)
	repoFilter := gcrcleaner.NewRepoFilter(regexRepo)

	imageFilter := gcrcleaner.NewImageFilter(*ageInHours)

	repoAPI := &gcrcleaner.RepoAPIWrapper{Context: ctx, Auth: auth}

	imageAPI := &gcrcleaner.ImageAPIWrapper{Context: ctx, Auth: auth}

	gcrCleaner := gcrcleaner.New(repoAPI, imageAPI, repoFilter, imageFilter)

	allSucceeded, err := gcrCleaner.Run(*repository, !(*dryRun))
	if err != nil {
		log.Fatalf("GCR cleaner error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}
