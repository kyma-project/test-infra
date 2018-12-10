package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kyma-project/test-infra/development/tools/pkg/dnscleaner"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

var (
	project = flag.String("project", "", "Project ID")
	dnsZone = flag.String("dns-zone", "", "Name of the zone in DNS")
	// dryRun  = flag.Bool("dry-run", true, "Dry Run enabled")
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
	cleaner := dnscleaner.NewCleaner(computeAPI, dnsAPI)
	cleaner.Run(*project, *dnsZone)

}
