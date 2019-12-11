package cluster

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// APIWrapper wraps the GCP api
type APIWrapper struct {
	ProjectID      string
	LocationID     string
	ClusterService *container.ProjectsLocationsClustersService
}

// NewWrapper creates a new cluster api based on the wrapper struct
func NewWrapper(ctx context.Context, opts Option) *APIWrapper {
	containerService, err := container.NewService(ctx, option.WithServiceAccountFile(opts.ServiceAccount))
	if err != nil {
		log.Fatalf("Couldn't create service handle for GCP: %w", err)
	}
	clusterService := containerService.Projects.Locations.Clusters
	return &APIWrapper{
		ClusterService: clusterService,
	}
}

// Create calls the wrapped GCP api to create a cluster
func (caw *APIWrapper) Create(ctx context.Context, name string, labels ...string) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", caw.ProjectID, caw.LocationID)

	ccRequest := &container.CreateClusterRequest{Cluster: &container.Cluster{
		Name: "test",
	}}

	createRequest, err := caw.ClusterService.Create(parent, ccRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("couldn't create cluster: %w", err)
	}
	fmt.Println(createRequest)
	return nil
}

// Delete calls the wrapped GCP api to delete a cluster
func (caw *APIWrapper) Delete(ctx context.Context, name string) error {
	return nil
}
