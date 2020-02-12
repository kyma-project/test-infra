package storage

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io/ioutil"

	gcs "cloud.google.com/go/storage"
)

// APIWrapper wraps the GCP api
type APIWrapper struct {
	ProjectID  string
	LocationID string
	GCSClient  *gcs.Client
}

func NewClient(ctx context.Context, opts Option, credentials string) (*Client, error) {
	gcsClient, err := gcs.NewClient(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("GCS client creation error %w", err)
	}

	api := &APIWrapper{
		ProjectID:  opts.ProjectID,
		LocationID: opts.LocationID,
		GCSClient:  gcsClient,
	}

	if client, err := New(opts, api); err != nil {
		return nil, fmt.Errorf("bucket client creation error %w", err)
	} else {
		return client, nil
	}
}

// CreateBucket attempts to create a storage bucket
func (caw *APIWrapper) CreateBucket(ctx context.Context, name string) error {
	attrs := &gcs.BucketAttrs{
		Name: name,
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
