package main

import (
	"context"
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
)

var (
	projectID  = flag.String("proj", "", "ProjectID of the GCP project [Required]")
	bucketName = flag.String("bucket", "", "Name of the storage bucket that contains the key [Required]")
	locationID = flag.String("loc", "global", "Location of the keyring used for encryption/decryption [Optional]")
	prefix     = flag.String("prefix", "", "Prefix for naming resources [Optional]")
)

func main() {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatalf("Requires the environment variable GOOGLE_APPLICATION_CREDENTIALS to be set to a GCP service account file.")
	}

	flag.Parse()
	if *projectID == "" {
		log.Fatalf("Missing required argument : -proj")
	}
	if *bucketName == "" {
		log.Fatalf("Missing required argument : -bucket")
	}

	ctx := context.Background()
	client, err := storage.New(ctx, storage.Option{Prefix: *prefix, ProjectID: *projectID, LocationID: *locationID})
	if err != nil {
		log.Errorf("Could not create GCS Storage Client: %v", err)
		os.Exit(1)
	}

	err = client.CreateBucket(*bucketName)
	if err != nil {
		log.Fatalf("Creating bucket failed: %v", err)
	}

	err = client.DeleteBucket(*bucketName)
	if err != nil {
		log.Fatalf("Deleting bucket failed: %v", err)
	}
}
