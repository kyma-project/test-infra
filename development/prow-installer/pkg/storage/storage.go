package storage

import (
	"context"
	"fmt"
)

// Option wrapper for relevant Options for the client
type Option struct {
	Prefix    string // storage prefix
	ProjectID string // GCP project ID
}

//go:generate mockery -name=API -output=automock -outpkg=automock -case=underscore

// Client wrapper for GCS storage API
type Client struct {
	Option
	api API
}

// Struct that defines each bucket to create
type Bucket struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location,omitempty"`
}

// API provides a mockable interface for the GCP api. Find the implementation of the GCP wrapped API in wrapped.go
type API interface {
	CreateBucket(ctx context.Context, name string, location string) error
	DeleteBucket(ctx context.Context, name string) error
	Read(ctx context.Context, bucket, storageObject string) ([]byte, error)
	Write(ctx context.Context, data []byte, bucket, storageObject string) error
}

// New returns a new Client, wrapping gcs for storage management on GCP
func New(projectID, prefix string, api API) (*Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required to initialize a client")
	}
	if api == nil {
		return nil, fmt.Errorf("api is required to initialize a client")
	}
	opts := Option{ProjectID: projectID, Prefix: prefix}
	return &Client{Option: opts, api: api}, nil
}

// CreateBucket attempts to create a storage bucket
func (sc *Client) CreateBucket(ctx context.Context, name string, location string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if sc.Prefix != "" {
		name = fmt.Sprintf("%s-%s", sc.Prefix, name)
	}
	return sc.api.CreateBucket(ctx, name, location)
}

// DeleteBucket attempts to delete a storage bucket
func (sc *Client) DeleteBucket(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return sc.api.DeleteBucket(ctx, name)
}

// Read attempts to read from a storage bucket
func (sc *Client) Read(ctx context.Context, bucket, storageObject string) ([]byte, error) {
	if bucket == "" {
		return nil, fmt.Errorf("bucket cannot be empty")
	}
	if storageObject == "" {
		return nil, fmt.Errorf("storageObject cannot be empty")
	}
	return sc.api.Read(ctx, bucket, storageObject)
}

// Write attempts to write to a storage bucket
func (sc *Client) Write(ctx context.Context, data []byte, bucket, storageObject string) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot write zero data")
	}
	if bucket == "" {
		return fmt.Errorf("bucket cannot be empty")
	}
	if storageObject == "" {
		return fmt.Errorf("storageObject cannot be empty")
	}
	return sc.api.Write(ctx, data, bucket, storageObject)
}

// WithPrefix modifies option to have a prefix
func (o Option) WithPrefix(pre string) Option {
	o.Prefix = pre
	return o
}

// WithProjectID modifies option to have a project id
func (o Option) WithProjectID(pid string) Option {
	o.ProjectID = pid
	return o
}
