package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/dnscleaner"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

const defaultAddressRegexpList = "(remoteenvs-)?gkeint-pr-.*,gke-upgrade-pr-.*"

var (
	project              = flag.String("project", "", "Project ID [Required]")
	dnsZone              = flag.String("dnsZone", "", "Name of the zone in DNS [Required]")
	dryRun               = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	addressNameRegexList = flag.String("addressRegexpList", defaultAddressRegexpList, "Address name regexp list. Separate items with commas. Matching addresses are considered for removal.")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *dnsZone == "" {
		fmt.Fprintln(os.Stderr, "missing -dnsZone flag")
		flag.Usage()
		os.Exit(2)
	}

	patterns := strings.Split(*addressNameRegexList, ",")
	regexpList := []*regexp.Regexp{}
	for _, pattern := range patterns {
		regexpList = append(regexpList, regexp.MustCompile(pattern))
	}
	ctx := context.Background()

	computeConn, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	computeSvc, err := compute.New(computeConn)
	if err != nil {
		log.Fatalf("Could not initialize gke client for Compute API: %v", err)
	}

	dnsSvc, err := dns.New(computeConn)
	if err != nil {
		log.Fatalf("Could not initialize gke client for Compute API: %v", err)
	}

	computeAPI := &dnscleaner.ComputeServiceWrapper{Context: ctx, Compute: computeSvc}
	dnsAPI := &dnscleaner.DNSServiceWrapper{Context: ctx, DNS: dnsSvc}
	shouldRemoveFunc := dnscleaner.DefaultIPAddressRemovalPredicate(regexpList, 1)

	cleaner := dnscleaner.NewCleaner(computeAPI, dnsAPI, shouldRemoveFunc)
	//TODO: use actual dryRun flag value
	cleaner.Run(*project, *dnsZone, false)
}
