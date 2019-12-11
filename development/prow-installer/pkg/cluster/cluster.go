package cluster

import (
	"context"
	"fmt"
)

type Option struct {
	ProjectID      string // GCP project ID
	ZoneID         string // zone of the cluster
	ServiceAccount string // filename of the serviceaccount to use
}

//go:generate mockery -name=API -output=automock -outpkg=automock -case=underscore

// Client wrapper for KMS and GCS secret storage
type Client struct {
	Option
	api API
}

// API provides a mockable interface for the GCP api. Find the implementation of the GCP wrapped API in wrapped.go
type API interface {
	Create(ctx context.Context, name string, labels map[string]string) error
	Delete(ctx context.Context, name string) error
}

// New returns a new Client, wrapping gke
func New(opts Option, api API) (*Client, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("ProjectID is required to initialize a client")
	}
	if opts.ZoneID == "" {
		return nil, fmt.Errorf("ZoneID is required to initialize a client")
	}
	if opts.ServiceAccount == "" {
		return nil, fmt.Errorf("ServiceAccount is required to initialize a client")
	}

	return &Client{Option: opts, api: api}, nil
}

// Create attempts to create a GKE cluster
func (cc *Client) Create(ctx context.Context, name string, labels map[string]string) error {
	return cc.api.Create(ctx, name, labels)
}

// Delete attempts to delete a GKE cluster
func (cc *Client) Delete(ctx context.Context, name string) error {
	return cc.api.Delete(ctx, name)
}
