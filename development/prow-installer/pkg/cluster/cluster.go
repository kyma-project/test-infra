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
	Create(ctx context.Context, name string, labels map[string]string, minPoolSize int, autoScaling bool) error
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
	if api == nil {
		return nil, fmt.Errorf("api is required to initialize a client")
	}
	return &Client{Option: opts, api: api}, nil
}

// Create attempts to create a GKE cluster
func (cc *Client) Create(ctx context.Context, name string, labels map[string]string, minPoolSize int, autoScaling bool) error {
	if minPoolSize < 1 {
		return fmt.Errorf("could not create cluster, minPoolSize should be > 0")
	}
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return cc.api.Create(ctx, name, labels, minPoolSize, autoScaling)
}

// Delete attempts to delete a GKE cluster
func (cc *Client) Delete(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return cc.api.Delete(ctx, name)
}

// WithProjectID modifies option to have a project id
func (o Option) WithProjectID(pid string) Option {
	o.ProjectID = pid
	return o
}

// WithZoneID modifies option to have a zone id
func (o Option) WithZoneID(z string) Option {
	o.ZoneID = z
	return o
}

// WithServiceAccount modifies option to have a service account
func (o Option) WithServiceAccount(sa string) Option {
	o.ServiceAccount = sa
	return o
}
