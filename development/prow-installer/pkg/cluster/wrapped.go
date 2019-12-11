package cluster

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	container "google.golang.org/api/container/v1"
	// "google.golang.org/api/option"
)

// APIWrapper wraps the GCP api
type APIWrapper struct {
	ProjectID      string
	ZoneID         string
	ClusterService *container.ProjectsZonesClustersService
}

// Create calls the wrapped GCP api to create a cluster
func (caw *APIWrapper) Create(ctx context.Context, name string, labels map[string]string, minPoolSize int, autoScaling bool) error {
	var pool []*container.NodePool

	pr, err := newNodePool(fmt.Sprintf("%s-pool", name), minPoolSize, autoScaling)
	if err != nil {
		return fmt.Errorf("couldn't define node pool for cluster: %w", err)
	}
	pool = append(pool, pr)

	ccRequest := &container.CreateClusterRequest{Cluster: &container.Cluster{
		Name:           name,
		ResourceLabels: labels,
		NodePools:      pool,
	}}

	createResponse, err := caw.ClusterService.Create(caw.ProjectID, caw.ZoneID, ccRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("couldn't create cluster: %w", err)
	}
	log.Printf("%#v\n", createResponse)
	return nil
}

// Delete calls the wrapped GCP api to delete a cluster
func (caw *APIWrapper) Delete(ctx context.Context, name string) error {
	deleteResponse, err := caw.ClusterService.Delete(caw.ProjectID, caw.ZoneID, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("couldn't delete cluster: %w", err)
	}
	log.Printf("%#v\n", deleteResponse)
	return nil
}

func newNodePool(name string, initialNodeCount int, autoScaling bool) (*container.NodePool, error) {
	scaling := &container.NodePoolAutoscaling{
		Enabled: autoScaling,
	}
	pool := &container.NodePool{
		Name:             name,
		InitialNodeCount: 1,
		Autoscaling:      scaling,
	}

	return pool, nil
}
