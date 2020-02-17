package cluster

import (
	"context"
	"fmt"
	"google.golang.org/api/container/v1"
)

type MockAPI struct {
}

func (api *MockAPI) Create(ctx context.Context, clusterConfig Cluster) error {
	var nodePools []*container.NodePool

	for _, pool := range clusterConfig.Pools {
		if nodePool, err := NewNodePool(pool); err != nil {
			return fmt.Errorf("error creating node pool configuration %w", err)
		} else {
			nodePools = append(nodePools, nodePool)
		}
	}

	_ = &container.CreateClusterRequest{
		Cluster: &container.Cluster{
			Name:           clusterConfig.Name,
			ResourceLabels: clusterConfig.Labels,
			Description:    clusterConfig.Description,
			NodePools:      nodePools,
		}}
	return nil
}

func (api *MockAPI) Delete(ctx context.Context, name string) error {
	return nil
}
