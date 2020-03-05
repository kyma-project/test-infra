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
	prefix     = flag.String("prefix", "", "Prefix for naming resources [Optional]")
	location   = flag.String("location", "", "Location of a bucket. Default US [Optional]")
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

	//storage service create
	storageService, err := storage.NewService(ctx, *projectID)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}
	client, err := storage.New(*projectID, *prefix, storageService)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}

	if err != nil {
		log.Fatalf("Could not create GCS Storage Client: %v", err)
	}

	err = client.CreateBucket(ctx, *bucketName, *location)
	if err != nil {
		log.Fatalf("Creating bucket failed: %v", err)
	}

	err = client.DeleteBucket(ctx, *bucketName)
	if err != nil {
		log.Fatalf("Deleting bucket failed: %v", err)
	}
}
