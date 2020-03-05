package cluster

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/container/v1"
)

type MockAPI struct {
}

func (api *MockAPI) Create(ctx context.Context, clusterConfig Cluster) error {
	var nodePools []*container.NodePool

	for _, pool := range clusterConfig.Pools {
		if nodePool, err := newNodePool(pool); err != nil {
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
	log.Printf("Method Create() called with arguments: %v, %v", ctx, clusterConfig)
	return nil
}

func (api *MockAPI) Delete(ctx context.Context, name string, zoneId string) error {
	log.Printf("Method Delete() called with arguments: %v, %v, %v", ctx, name, zoneId)
	return nil
}
