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
	"github.com/kyma-project/test-infra/development/tools/pkg/dnscleaner"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

var (
	project     = flag.String("project", "", "Project ID [Required]")
	zone        = flag.String("zone", "", "zone name [Required]")
	name        = flag.String("name", "", "DNS resource name [Required]")
	address     = flag.String("address", "", "DNS resource's attached IP [Required]")
	rtype       = flag.String("type", "A", "DNS Type to search for (default: \"A\"")
	ttl         = flag.Int64("ttl", 300, "TTL of the resource to search for (default: 300)")
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

	common.ShoutFirst("Running with arguments: project: \"%s\", dryRun: %t, dns resource name: \"%s\"", *project, *dryRun, *name)
	ctx := context.Background()

	connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	dnsSvc, err := dns.NewService(ctx, option.WithHTTPClient(connection))
	if err != nil {
		log.Fatalf("Could not initialize dns API client: %v", err)
	}

	dnsAPI := &dnscleaner.DNSAPIWrapper{Service: dnsSvc}

	der := dnscleaner.New(dnsAPI, *maxAttempts, *backoff, !(*dryRun))

	entryRemoverErr := der.Run(*project, *zone, *name, *address, *rtype, *ttl)

	if entryRemoverErr != nil {
		log.Warn("Operation failed.")
		log.Fatalf("DNS Cleaner error: %v", entryRemoverErr)
	}

	common.Shout("Finished")
}
