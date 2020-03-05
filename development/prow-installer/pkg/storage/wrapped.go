package storage

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io/ioutil"
	"os"

	gcs "cloud.google.com/go/storage"
)

// APIWrapper wraps the GCP api
type APIWrapper struct {
	ProjectID string
	GCSClient *gcs.Client
}

func NewService(ctx context.Context, projectID string) (*APIWrapper, error) {
	gcsClient, err := gcs.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		return nil, err
	}
	api := &APIWrapper{
		ProjectID: projectID,
		GCSClient: gcsClient,
	}
	return api, nil
}

// CreateBucket attempts to create a storage bucket
func (caw *APIWrapper) CreateBucket(ctx context.Context, name string, location string) error {
	attrs := &gcs.BucketAttrs{
		Location: location,
		Name:     name,
	}

	err := caw.GCSClient.Bucket(name).Create(ctx, caw.ProjectID, attrs)
	if err != nil {
		return fmt.Errorf("Error creating the bucket: %w", err)
	}
	return nil
}

// DeleteBucket attempts to delete a storage bucket
func (caw *APIWrapper) DeleteBucket(ctx context.Context, name string) error {
	err := caw.GCSClient.Bucket(name).Delete(ctx)
	if err != nil {
		return fmt.Errorf("Error deleting the bucket: %w", err)
	}
	return nil
}

// Read attempts to read from a storage bucket
func (caw *APIWrapper) Read(ctx context.Context, bucket, storageObject string) ([]byte, error) {
	rc, err := caw.GCSClient.Bucket(bucket).Object(storageObject).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Creating bucket reader failed: %w", err)
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("Reading from bucket failed: %w", err)
	}
	return data, nil
}

// Write attempts to write to a storage bucket
func (caw *APIWrapper) Write(ctx context.Context, data []byte, bucket, storageObject string) error {
	wc := caw.GCSClient.Bucket(bucket).Object(storageObject).NewWriter(ctx)
	if _, err := wc.Write(data); err != nil {
		return fmt.Errorf("Writing to bucket failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
