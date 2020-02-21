package cluster

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"os"
	"os/exec"
	"time"

	container "google.golang.org/api/container/v1"
	// "google.golang.org/api/option"
)

// APIWrapper wraps the GCP api
type APIWrapper struct {
	ProjectID      string
	ClusterService *container.ProjectsZonesClustersService
}

func NewClient(ctx context.Context, opts Option, credentials string) (*Client, error) {
	containerService, err := container.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("container service creation error %w", err)
	}
	api := &APIWrapper{
		ProjectID:      opts.ProjectID,
		ClusterService: containerService.Projects.Zones.Clusters,
	}

	if client, err := New(opts, api); err != nil {
		return nil, fmt.Errorf("cluster client creation error %w", err)
	} else {
		return client, nil
	}
}

// Create calls the wrapped GCP api to create a cluster
func (caw *APIWrapper) Create(ctx context.Context, clusterConfig Cluster) (string, error) {
	var nodePools []*container.NodePool

	for _, pool := range clusterConfig.Pools {
		if nodePool, err := NewNodePool(pool); err != nil {
			return "", fmt.Errorf("error creating node pool configuration: %w", err)
		} else {
			nodePools = append(nodePools, nodePool)
		}
	}

	ccRequest := &container.CreateClusterRequest{
		Cluster: &container.Cluster{
			Name:           clusterConfig.Name,
			ResourceLabels: clusterConfig.Labels,
			Description:    clusterConfig.Description,
			NodePools:      nodePools,
		}}

	_, err := caw.ClusterService.Create(caw.ProjectID, clusterConfig.Location, ccRequest).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("couldn't create cluster: %w", err)
	} else {
		log.Infof("successfully created cluster %s", clusterConfig.Name)
	}

	// get kubeconfig
	kubeconfig := "./.kube/" + clusterConfig.Name + "_config"

	for {
		cmd := exec.Command("gcloud", "container", "clusters", "get-credentials", fmt.Sprintf(clusterConfig.Name),
			"--project", caw.ProjectID,
			"--zone", clusterConfig.Location)
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfig))

		// TODO handle gcloud errors specifically
		if err := cmd.Run(); err != nil {
			log.Errorf("error getting kubeconfig %v. Retrying in 5 seconds...", err)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}
	return kubeconfig, nil
}

// Delete calls the wrapped GCP api to delete a cluster
func (caw *APIWrapper) Delete(ctx context.Context, name string, zoneId string) error {
	deleteResponse, err := caw.ClusterService.Delete(caw.ProjectID, zoneId, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("couldn't delete cluster: %w", err)
	}
	log.Printf("%#v\n", deleteResponse)
	return nil
}

func NewNodePool(nodePool Pool) (*container.NodePool, error) {
	if nodePool.Size == 0 {
		return nil, fmt.Errorf("size must be at least 1")
	}
	pool := &container.NodePool{
		Name:             nodePool.Name,
		InitialNodeCount: nodePool.Size,
		Autoscaling: &container.NodePoolAutoscaling{
			Enabled:      nodePool.Autoscaling.Enabled,
			MaxNodeCount: nodePool.Autoscaling.MaxNodeCount,
			MinNodeCount: nodePool.Autoscaling.MinNodeCount,
		},
		Config: &container.NodeConfig{
			DiskSizeGb:  nodePool.NodeConfig.DiskSizeGb,
			DiskType:    nodePool.NodeConfig.DiskType,
			MachineType: nodePool.NodeConfig.MachineType,
		},
	}

	return pool, nil
}
