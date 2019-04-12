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
	dnscleaner "github.com/kyma-project/test-infra/development/tools/pkg/longlastingdnscleaner"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/dns/v1"
)

var (
	project     = flag.String("project", "", "Project ID [Required]")
	zone        = flag.String("zone", "", "zone name [Required]")
	name        = flag.String("name", "", "DNS resource name [Required]")
	address     = flag.String("address", "", "DNS resource's attached IP [Required]")
	rtype       = flag.String("type", "A", "DNS Type to search for (default: \"A\"")
	ttl         = flag.Int64("ttl", 300, "TTL of the resource to search for (default: 300)")
	maxAttempts = flag.Uint("attempts", 3, "Maximal number of attempts until scripts stops trying to delete IP (default: 3)")
	timeout     = flag.Uint("timeout", 5, "Timeout in seconds, will increase over time to reduce API calls (default: 5)")
	dryRun      = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *name == "" {
		fmt.Fprintln(os.Stderr, "missing -name flag")
		flag.Usage()
		os.Exit(2)
	}

	if *address == "" {
		fmt.Fprintln(os.Stderr, "missing -address flag")
		flag.Usage()
		os.Exit(2)
	}

	if *zone == "" {
		fmt.Fprintln(os.Stderr, "missing -zone flag")
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, ipname: \"%s\"", *project, *dryRun, *name)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	dnsSvc, err := dns.New(connection)
	if err != nil {
		log.Fatalf("Could not initialize dns API client: %v", err)
	}

	dnsAPI := &dnscleaner.DNSAPIWrapper{Context: ctx, Service: dnsSvc}

	der := dnscleaner.NewDNSEntryRemover(dnsAPI)

	success, err := der.Run(*project, *zone, *name, *address, *rtype, *ttl, *maxAttempts, *timeout, !(*dryRun))

	if err != nil {
		log.Fatalf("Cluster collector error: %v", err)
	}

	if !success {
		log.Warn("Operation failed.")
	}

	common.Shout("Finished")
}
