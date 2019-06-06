package main

import (
	"flag"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/test-infra/development/tools/pkg/vulnerabilitycollector"
)

var (
	project = flag.String("project", "", "Project ID [Required]")
	url     = flag.String("url", "", "Resource Url [Required]")
)

func main() {
	flag.Parse()

	vs, error := vulnerabilitycollector.FindVulnerabilityOccurrencesForImage(*url, *project)
	if error != nil {
		log.Fatalf("Could not get authenticated client: %v", error)
	}
	var count int
	var warn int
	var packages map[string]string
	for _, element := range vs {

		req := element.GetVulnerability()
		packages = make(map[string]string)
		if strings.Contains("HIGH", req.Severity.String()) {

			log.Warn("Severity ", req.Severity, " ", req.PackageIssue[0].AffectedLocation.Package, " ", req.PackageIssue[0].AffectedLocation.Version.Name)
			packages[req.PackageIssue[0].AffectedLocation.Package] = req.PackageIssue[0].AffectedLocation.Version.Name
			warn++
		}
		count++
	}
	if warn > 0 {
		log.Warn("Number of High issues ", warn)
	}

	log.Infof("Number of issues %d", count)

}
