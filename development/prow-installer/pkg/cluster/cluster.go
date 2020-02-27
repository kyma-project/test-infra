package cluster

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
	"k8s.io/client-go/kubernetes"
)

type Option struct {
	Prefix         string // global prefix
	ProjectID      string // GCP project ID
	ServiceAccount string // filename of the serviceaccount to use
}

//go:generate mockery -name=API -output=automock -outpkg=automock -case=underscore

// Client wrapper for KMS and GCS secret storage
type Client struct {
	Option
	api API
}

type Cluster struct {
	Name                  string            `yaml:"name"`
	Location              string            `yaml:"location"`
	Description           string            `yaml:"description,omitempty"`
	Labels                map[string]string `yaml:"labels,omitempty"`
	Pools                 []Pool            `yaml:"pools"`
	InitialClusterVersion string            `yaml:"kubernetesVersion,omitempty"`
	K8sClient             *kubernetes.Clientset
	Populator             *k8s.Populator
}

// node pool settings
type Pool struct {
	Name        string      `yaml:"name"`
	Size        int64       `yaml:"initialSize,omitempty"`
	Autoscaling Autoscaling `yaml:"autoscaling,omitempty"`
	NodeConfig  NodeConfig  `yaml:"config,omitempty"`
}

type NodeConfig struct {
	MachineType string `yaml:"machineType,omitempty"`
	DiskType    string `yaml:"diskType,omitempty"`
	DiskSizeGb  int64  `yaml:"diskSizeGb,omitempty"`
}

// Autoscaling features for cluster
type Autoscaling struct {
	Enabled      bool  `yaml:"enabled"`
	MinNodeCount int64 `yaml:"minNodeCount"`
	MaxNodeCount int64 `yaml:"maxNodeCount"`
}

// API provides a mockable interface for the GCP api. Find the implementation of the GCP wrapped API in wrapped.go
type API interface {
	Create(ctx context.Context, clusterConfig Cluster) error
	Delete(ctx context.Context, name string, zoneID string) error
}

// New returns a new Client, wrapping gke
func New(opts Option, api API) (*Client, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("ProjectID is required to initialize a client")
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
func (cc *Client) Create(ctx context.Context, clusterConfig Cluster) error {
	if clusterConfig.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if cc.Prefix != "" {
		clusterConfig.Name = fmt.Sprintf("%s-%s", cc.Prefix, clusterConfig.Name)
	}
	return cc.api.Create(ctx, clusterConfig)
}

// Delete attempts to delete a GKE cluster
func (cc *Client) Delete(ctx context.Context, name string, zoneId string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if zoneId == "" {
		return fmt.Errorf("zoneId cannot be empty")
	}
	return cc.api.Delete(ctx, name, zoneId)
}

// WithProjectID modifies option to have a project id
func (o Option) WithProjectID(pid string) Option {
	o.ProjectID = pid
	return o
}

// WithServiceAccount modifies option to have a service account
func (o Option) WithServiceAccount(sa string) Option {
	o.ServiceAccount = sa
	return o
}
