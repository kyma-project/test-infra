package main

import (
	"context"
	flag "github.com/spf13/pflag"
	"log"
	"os"

	"github.com/kyma-project/test-infra/development/gcp/pkg/dnsclient"
)

var (
	zoneName        = flag.StringP("zone", "z", "build-kyma-workloads", "DNS zone name [Optional]")
	name            = flag.StringP("name", "n", "", "Record name example: www.example.com. [Required]")
	data            = flag.StringP("data", "d", "", "Record data [Required]")
	project         = flag.StringP("project", "p", "sap-kyma-prow-workloads", "GCP project name [Required]")
	recordType      = flag.StringP("type", "t", "A", "DNS record type [Optional]")
	ttl             = flag.Int64P("ttl", "T", 1800, "Record time to live [Optional]")
	credentialsfile = flag.StringP("credentialsfile", "c", "", "Google Application Credentials file path. [Required]")
	deleteRecord    = flag.BoolP("delete", "D", false, "If set to true, record will be deleted. Default: false [Optional]")
)

func main() {
	flag.Parse()
	flags := []*string{name, data, project}
	for _, val := range flags {
		if *val == "" {
			log.Fatalf("Required arguments must be provided.")
		}
	}
	if *credentialsfile != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *credentialsfile); err != nil {
			log.Fatalf("Error when setting environment variable GOOGLE_APPLICATION_CREDENTIALS.")
		}
	} else if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) < 1 {
		log.Fatalf("Google application credentials are required")
	}

	ctx := context.Background()
	opts := dnsclient.NewRecordOpts(*project, *zoneName, *name, *data, *recordType, *ttl)

	service, err := dnsclient.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed get dns service: %v", err)
	}

	client, err := dnsclient.New(service)
	if err != nil {
		log.Fatalf("failed craete dns client: %v", err)
	}

	var change *dnsclient.DNSChange
	switch *deleteRecord {
	case false:
		change = client.NewDNSChange(opts).AddRecord()
	case true:
		change = client.NewDNSChange(opts).DeleteRecord()
	}
	_, err = client.DoChange(ctx, change)
	if err != nil {
		log.Fatalf("error when changin record: %v", err)
	}

	switch *deleteRecord {
	case false:
		log.Printf("added record %s to zone %s", *name, *zoneName)
	case true:
		log.Printf("deleted record %s from zone %s", *name, *zoneName)
	}
}
