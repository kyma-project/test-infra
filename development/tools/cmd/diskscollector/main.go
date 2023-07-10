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
	"github.com/kyma-project/test-infra/development/tools/pkg/diskscollector"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

const defaultDiskNameRegex = "^gke-gkeint.*[-]pvc[-]"

var (
	project       = flag.String("project", "", "Project ID [Required]")
	dryRun        = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	ageInHours    = flag.Int("ageInHours", 2, "Disk age in hours. Disks older than: now()-ageInHours are considered for removal.")
	diskNameRegex = flag.String("diskNameRegex", defaultDiskNameRegex, "Disk name regex. Matching disks are considered for removal.")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ageInHours: %d, diskNameRegex: \"%s\"", *project, *dryRun, *ageInHours, *diskNameRegex)
	context := context.Background()

	connection, err := google.DefaultClient(context, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := compute.NewService(context, option.WithHTTPClient(connection))
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	rx := regexp.MustCompile(*diskNameRegex)

	zoneAPI := &diskscollector.ZoneAPIWrapper{Context: context, Service: svc.Zones}
	diskAPI := &diskscollector.DiskAPIWrapper{Context: context, Service: svc.Disks}
	filter := diskscollector.NewDiskFilter(rx, *ageInHours)

	gc := diskscollector.NewDisksGarbageCollector(zoneAPI, diskAPI, filter)

	allSucceeded, err := gc.Run(*project, !(*dryRun))
	if err != nil {
		log.Fatalf("Disk collector error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}
