package storage

import (
	"context"
	"fmt"
	"io/ioutil"

	gcs "cloud.google.com/go/storage"
)

// Option wrapper for relevant Options for the client
type Option struct {
	Prefix     string // storage prefix
	ProjectID  string // GCP project ID
	LocationID string // location of the key rings
}

// Client wrapper for GCS storage API
type Client struct {
	Option
	ctx context.Context
}

// New returns a new Client, wrapping gcs for storage management on GCP
func New(ctx context.Context, opts Option) (*Client, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("ProjectID is required to initialize a client")
	}
	if opts.LocationID == "" {
		return nil, fmt.Errorf("LocationID is required to initialize a client")
	}
	return &Client{Option: opts, ctx: ctx}, nil
}

func (sc *Client) CreateBucket(name string) error {
	client, err := gcs.NewClient(sc.ctx)
	if err != nil {
		return fmt.Errorf("Initializing storage client failed: %w", err)
	}

	attrs := &gcs.BucketAttrs{
		Name: name,
	}

	err = client.Bucket(name).Create(sc.ctx, sc.ProjectID, attrs)
	if err != nil {
		return fmt.Errorf("Error creating the bucket: %w", err)
	}
	return nil
}

func (sc *Client) DeleteBucket(name string) error {
	client, err := gcs.NewClient(sc.ctx)
	if err != nil {
		return fmt.Errorf("Initializing storage client failed: %w", err)
	}

	err = client.Bucket(name).Delete(sc.ctx)
	if err != nil {
		return fmt.Errorf("Error deleting the bucket: %w", err)
	}
	return nil
}

func (sc *Client) Read(bucket, storageObject string) ([]byte, error) {
	client, err := gcs.NewClient(sc.ctx)
	if err != nil {
		return nil, fmt.Errorf("Initializing storage client failed: %w", err)
	}

	rc, err := client.Bucket(bucket).Object(storageObject).NewReader(sc.ctx)
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

func (sc *Client) Write(data []byte, bucket, storageObject string) error {
	client, err := gcs.NewClient(sc.ctx)
	if err != nil {
		return fmt.Errorf("Initializing storage client failed: %w", err)
	}

	wc := client.Bucket(bucket).Object(storageObject).NewWriter(sc.ctx)
	if _, err := wc.Write(data); err != nil {
		return fmt.Errorf("Writing to bucket failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}
