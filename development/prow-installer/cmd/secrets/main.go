package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	gcs "cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/secrets"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
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

	kmsClient, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Initialising KMS client failed: %w", err))
	}

	wrappedKmsAPI := &secrets.APIWrapper{
		ProjectID:  *projectID,
		LocationID: *locationID,
		KmsRing:    *kmsRing,
		KmsKey:     *kmsKey,
		KmsClient:  kmsClient,
	}

	kmsClientOpts := secrets.Option{}
	kmsClientOpts = kmsClientOpts.WithProjectID(*projectID).WithLocationID(*locationID).WithKmsRing(*kmsRing).WithKmsKey(*kmsKey).WithServiceAccount(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	secretClient, err := secrets.New(kmsClientOpts, wrappedKmsAPI)
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Could not create KMS Management Client: %w", err))
	}

	gcsClient, err := gcs.NewClient(ctx)
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Initializing storage client failed: %w", err))
	}

	wrappedGCSAPI := &storage.APIWrapper{
		ProjectID: *projectID,
		GCSClient: gcsClient,
	}

	gcsClientOpts := storage.Option{}
	gcsClientOpts = gcsClientOpts.WithPrefix(*prefix).WithProjectID(*projectID).WithServiceAccount(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	storageClient, err := storage.New(gcsClientOpts, wrappedGCSAPI)
	if err != nil {
		log.Fatalf("Could not create GCS Storage Client: %v", err)
	}

	data, err := secretClient.Encrypt(ctx, []byte("this is secret"))
	if err != nil {
		log.Fatalf("Encrypting secret failed: %v", err)
	}
	err = storageClient.Write(ctx, data, *bucketName, "secretName")
	if err != nil {
		log.Fatalf("Storing secret failed: %v", err)
	}

	b, err := storageClient.Read(ctx, *bucketName, "secretName")
	if err != nil {
		log.Fatalf("Reading secret failed: %v", err)
	}

	plain, err := secretClient.Decrypt(ctx, b)
	if err != nil {
		log.Fatalf("Decrypting secret failed: %v", err)
	}
	fmt.Println(string(plain))
}
