package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/development/gcp/pkg/dnsclient"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"time"

	//"log"
	"os"
)

var deleteRecord = flag.BoolP("delete", "D", false, "If set to true, record will be deleted. Default: false [Optional]")

func main() {
	flag.Parse()
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout
	if os.Getenv("CERTBOT_AUTH_OUTPUT") != "recordadded" && *deleteRecord {
		log.WithFields(logrus.Fields{"desc": "Delete record requested but record was not added earlier"}).Fatal("Exiting, nothing to do")
	}
	if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) < 1 {
		log.WithFields(logrus.Fields{"desc": "Google application credentials are required"}).Fatal("Error, terminating")
	}

	name := fmt.Sprintf("_acme-challenge.%s.", os.Getenv("CERTBOT_DOMAIN"))
	data := os.Getenv("CERTBOT_VALIDATION")
	recordType := "TXT"
	ctx := context.Background()
	opts := dnsclient.NewRecordOpts("", "", name, data, recordType, 1)

	service, err := dnsclient.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed get dns service: %v", err)
	}

	client, err := dnsclient.New(service)
	if err != nil {
		log.Fatalf("failed create dns client: %v", err)
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
		log.Fatalf("error when changing record: %v", err)
	}

	switch *deleteRecord {
	case false:
		fmt.Printf("%s", "recordadded\n")
		time.Sleep(61 * time.Second)
	case true:
		log.Info("Record deleted. DNS authentication finished.")
	}
}
