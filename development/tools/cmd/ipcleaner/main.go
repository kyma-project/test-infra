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

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/ipcleaner"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

// nat-auto-ip is *probably* from Gardener, let them handle removal
const defaultIpNameIgnoreRegex = "^nightly|weekly|nat-auto-ip"

var (
	project           = flag.String("project", "", "Project ID [Required]")
	dryRun            = flag.Bool("dry-run", true, "Dry Run enabled, nothing is deleted")
	ageInHours        = flag.Int("age-in-hours", 2, "Disk age in hours. Disks older than: now()-ageInHours are considered for removal.")
	ipNameIgnoreRegex = flag.String("disk-name-regex", defaultIpNameIgnoreRegex, "Ignored IP name regex. Matching IPs are not considered for removal.")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ageInHours: %d, ipNameIgnorRegex: \"%s\"", *project, *dryRun, *ageInHours, *ipNameIgnoreRegex)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	computeSvc, err := compute.New(connection)
	if err != nil {
		log.Fatalf("Could not initialize compute API client: %v", err)
	}

	rx := regexp.MustCompile(*ipNameIgnoreRegex)
	filter := ipcleaner.NewIpFilter(rx, *ageInHours)

	// TODO

	addressAPI := &ipcleaner.AddressAPIWrapper{Context: ctx, Service: computeSvc.Addresses}
	regionAPI := &ipcleaner.RegionAPIWrapper{Context: ctx, Service: computeSvc.Regions}

	ipr := ipcleaner.New(addressAPI, regionAPI, filter)

	// removalErr := ipr.Run(*project, *region, *ipName)

	allSucceeded, err := ipr.Run(*project, !(*dryRun))
	if err != nil {
		log.Fatalf("IP clenaer error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}
