package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/development/gcp/pkg/dnsclient"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
	"time"
)

var deleteRecord = flag.BoolP("delete", "D", false, "If set to true, record will be deleted. Default: false [Optional]")

func main() {
	flag.Parse()
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.SetLevel(logrus.InfoLevel)
	log.Out = os.Stdout
	if ! strings.Contains(os.Getenv("CERTBOT_AUTH_OUTPUT"), "status_message=dns_record_added") && *deleteRecord {
		log.WithFields(logrus.Fields{
			"topic": "dns record change",
			"action": "delete",
			"desc": "Delete record action requested, but record was not added earlier",
		}).Fatal("Nothing to do, exiting")
	}
	if len(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")) < 1 {
		log.WithFields(logrus.Fields{
			"topic": "GCP authentication",
			"desc": "GOOGLE_APPLICATION_CREDENTIALS env variable is required",
		}).Fatal("Not good, terminating")
	}

	name := fmt.Sprintf("_acme-challenge.%s.", os.Getenv("CERTBOT_DOMAIN"))
	data := os.Getenv("CERTBOT_VALIDATION")
	recordType := "TXT"
	ctx := context.Background()
	opts := dnsclient.NewRecordOpts("", "", name, data, recordType, 1)

	service, err := dnsclient.NewService(ctx)
	if err != nil {
		log.WithFields(logrus.Fields{
			"topic": "GCP dns client",
			"desc": "Failed to get dnsclient service instance",
			"error": fmt.Sprintf("%v", err),
		}).Fatal("Not good, terminating")
	}

	client, err := dnsclient.New(service)
	if err != nil {
		log.WithFields(logrus.Fields{
			"topic": "GCP dns client",
			"desc": "Failed to get dnsclient client instance",
			"error": fmt.Sprintf("%v", err),
		}).Fatal("Not good, terminating")
	}

	var change *dnsclient.DNSChange
	switch *deleteRecord {
	case false:
		change = client.NewDNSChange(opts).AddRecord()
		log.WithFields(logrus.Fields{
			"topic": "dns record change",
			"action": "add record",
			"domain": os.Getenv("CERTBOT_DOMAIN"),
			"record_name": name,
			"validation_string": os.Getenv("CERTBOT_VALIDATION"),
		}).Info("Record change requested")
	case true:
		change = client.NewDNSChange(opts).DeleteRecord()
		log.WithFields(logrus.Fields{
			"topic": "dns record change",
			"action": "delete record",
			"domain": os.Getenv("CERTBOT_DOMAIN"),
			"record_name": name,
		}).Info("Record change requested")
	}
	_, err = client.DoChange(ctx, change)
	if err != nil {
		log.WithFields(logrus.Fields{
			"topic": "dns record change",
			"desc": "Execution of requested change failed",
			"error": fmt.Sprintf("%v", err),
		}).Fatal("Not good, terminating")
	}

	switch *deleteRecord {
	case false:
		log.WithFields(logrus.Fields{
			"topic": "dns record change",
			"action": "add record",
			"domain": os.Getenv("CERTBOT_DOMAIN"),
			"record_name": name,
			"validation_string": os.Getenv("CERTBOT_VALIDATION"),
			"status_message": "dns_record_added",
		}).Info("Record added.")
		time.Sleep(61 * time.Second)
	case true:
		log.WithFields(logrus.Fields{
			"topic": "dns record change",
			"action": "delete record",
			"domain": os.Getenv("CERTBOT_DOMAIN"),
			"record_name": name,
		}).Info("Record deleted.")
	}
}
