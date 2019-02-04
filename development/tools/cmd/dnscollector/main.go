package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/dnscollector"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

const defaultAddressRegexpList = "(remoteenvs-)?gkeint-pr-.*,gke-upgrade-pr-.*"

var (
	project              = flag.String("project", "", "Project ID [Required]")
	dnsManagedZone       = flag.String("dnsZone", "", "Name of the DNS Managed Zone [Required]")
	dryRun               = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	ageInHours           = flag.Int("ageInHours", 2, "IP Address age in hours. Addresses older than: now()-ageInHours are considered for removal.")
	addressNameRegexList = flag.String("addressRegexpList", defaultAddressRegexpList, "Address name regexp list. Separate items with commas. Matching addresses are considered for removal.")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}

	if *dnsManagedZone == "" {
		fmt.Fprintln(os.Stderr, "missing -dnsZone flag")
		flag.Usage()
		os.Exit(2)
	}

	patterns := splitPatterns(*addressNameRegexList)
	regexpList := []*regexp.Regexp{}
	for _, pattern := range patterns {
		regexpList = append(regexpList, regexp.MustCompile(pattern))
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", dnsZone: \"%s\", dryRun: %t, ageInHours: %d, addressRegexpList: %s", *project, *dnsManagedZone, *dryRun, *ageInHours, quoteElems(patterns))
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

	computeAPI := &dnscollector.ComputeServiceWrapper{Context: ctx, Compute: computeSvc}
	dnsAPI := &dnscollector.DNSServiceWrapper{Context: ctx, DNS: dnsSvc}
	shouldRemoveFunc := dnscollector.DefaultIPAddressRemovalPredicate(regexpList, *ageInHours)

	cleaner := dnscollector.NewCleaner(computeAPI, dnsAPI, shouldRemoveFunc)
	makeChanges := !*dryRun
	cleaner.Run(*project, *dnsManagedZone, makeChanges)
}

func quoteElems(elems []string) string {

	fmt.Println(len(elems))
	res := "\"" + elems[0] + "\""
	for i := 1; i < len(elems); i++ {
		res = res + ","
		res = res + "\"" + elems[i] + "\""
	}

	return "[" + res + "]"
}

func splitPatterns(commaSeparated string) []string {

	res := []string{}
	values := strings.Split(commaSeparated, ",")
	for _, pattern := range values {
		res = append(res, pattern)
	}

	return res
}
