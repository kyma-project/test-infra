package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/secrets"
)

var (
	projectID  = flag.String("proj", "", "ProjectID of the GCP project [Required]")
	bucketName = flag.String("bucket", "", "Name of the storage bucket that contains the key [Required]")
	kmsRing    = flag.String("ring", "", "Key Ring name of the symmetric KMS key to use [Required]")
	kmsKey     = flag.String("key", "", "KMS Key mame from Key Ring to use [Required]")
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
	if *kmsRing == "" {
		log.Fatalf("Missing required argument : -ring")
	}
	if *kmsKey == "" {
		log.Fatalf("Missing required argument : -key")
	}

	ctx := context.Background()
	client, err := secrets.New(ctx, secrets.Option{Prefix: *prefix, ProjectID: *projectID, LocationID: *locationID, Bucket: *bucketName, KmsRing: *kmsRing, KmsKey: *kmsKey})
	if err != nil {
		log.Errorf("Could not create SecretClient: %v", err)
		os.Exit(1)
	}

	err = client.StoreSecret([]byte("this is secret"), "secretName")
	if err != nil {
		log.Fatalf("Storing secret failed: %v", err)
	}

	b, err := client.ReadSecret("secretName")
	if err != nil {
		log.Fatalf("Reading secret failed: %v", err)
	}
	fmt.Println(string(b))
}
