package clusterscollector

import (
	"context"
	"fmt"

	container "google.golang.org/api/container/v1"
)

// ClusterAPIWrapper abstracts GCP ProjectsLocationsClustersService API
type ClusterAPIWrapper struct {
	Context context.Context
	Service *container.ProjectsLocationsClustersService
}

// ListClusters delegates to ProjectsLocationsClustersService.List(parent) function
func (caw *ClusterAPIWrapper) ListClusters(project string) ([]*container.Cluster, error) {
	lcr, err := caw.Service.List(fmt.Sprintf("projects/%s/locations/-", project)).Context(caw.Context).Do()
	if err != nil {
		return nil, err
	}
	return lcr.Clusters, nil
}

// RemoveCluster delegates to ProjectsLocationsClustersService.Delete(name) function
func (caw *ClusterAPIWrapper) RemoveCluster(project, zone, name string) error {

	_, err := caw.Service.Delete(fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, zone, name)).Do()

	if err != nil {
		return err
	}

	return nil
}
