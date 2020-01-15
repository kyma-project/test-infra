package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/dnscollector"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
)

const defaultAddressRegexpList = "^((?!nightly|weekly).)*"
const minAgeInHours = 1
const minPatternLength = 5

var (
	project               = flag.String("project", "", "project id [required]")
	regions               = flag.String("regions", "", "comma-separted list of GCP regions [required]")
	dnsZone               = flag.String("dnsZone", "", "Name of the DNS Managed Zone [Required]")
	dryRun                = flag.Bool("dryRun", true, "Dry Run enabled, nothing is deleted")
	ageInHours            = flag.Int("ageInHours", 2, "IP Address age in hours. Addresses older than: now()-ageInHours are considered for removal.")
	addressNameRegexpList = flag.String("addressRegexpList", defaultAddressRegexpList, "Address name regexp list. Separate items with commas, spaces are trimmed. Matching addresses are considered for removal.")
)

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Fprint(os.Stderr, "missing -project flag\n\n")
		flag.Usage()
		os.Exit(2)
	}

	if *dnsZone == "" {
		fmt.Fprint(os.Stderr, "missing -dnsZone flag\n\n")
		flag.Usage()
		os.Exit(2)
	}

	regionsList := splitPatterns(*regions)
	for _, region := range regionsList {
		if len(region) == 0 {
			fmt.Fprint(os.Stderr, "invalid region: \"\"\n\n")
			flag.Usage()
			os.Exit(2)
		}
	}

	patterns := splitPatterns(*addressNameRegexpList)
	regexpList := []*regexp.Regexp{}
	for _, pattern := range patterns {
		if len(pattern) < minPatternLength {
			fmt.Fprintf(os.Stderr, "invalid pattern: \"%s\". Value must not be shorter than %d characters.\n\n", pattern, minPatternLength)
			flag.Usage()
			os.Exit(2)
		}
		regexpList = append(regexpList, regexp.MustCompile(pattern))
	}

	if *ageInHours < minAgeInHours {
		fmt.Fprintf(os.Stderr, "invalid ageInHours. Value must not be smaller than %d\n\n", minAgeInHours)
		flag.Usage()
		os.Exit(2)
	}

	common.ShoutFirst("Running with arguments: project: \"%s\", regions: \"%s\", dnsZone: \"%s\", dryRun: %t, ageInHours: %d, addressRegexpList: %s", *project, quoteElems(regionsList), *dnsZone, *dryRun, *ageInHours, quoteElems(patterns))
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

	cleaner := dnscollector.New(computeAPI, dnsAPI, shouldRemoveFunc)
	allSucceeded, err := cleaner.Run(*project, *dnsZone, regionsList, !(*dryRun))

	if err != nil {
		log.Fatalf("IP/DNS collector error: %v", err)
	}

	if !allSucceeded {
		log.Warn("Some operations failed.")
	}

	common.Shout("Finished")
}

func quoteElems(elems []string) string {

	res := "\"" + elems[0] + "\""
	for i := 1; i < len(elems); i++ {
		res = res + ",\"" + elems[i] + "\""
	}

	return "[" + res + "]"
}

func splitPatterns(commaSeparated string) []string {

	res := []string{}
	values := strings.Split(commaSeparated, ",")
	for _, pattern := range values {
		val := strings.Trim(pattern, " ")
		res = append(res, val)
	}

	return res
}
