package k8s


import (
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"log"
	"os"
	"time"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)


type API interface {
	Get(ctx context.Context, clusterID string) (*container.Cluster, error)
}

type Client struct {
	K8sclient *kubernetes.Clientset
}

// Refactor prow-installer cluster package client implementation to get rid of this method. prow-installer package should be able to provide client for API interface implemented here.

func NewGKEClient (ctx context.Context, projectID string, zoneID string) (*cluster.APIWrapper, error) {
	containerService, err := container.NewService(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("failed creating gke client, got: %v", err)
	}

	api := &cluster.APIWrapper{
		ProjectID:      projectID,
		ZoneID:         zoneID,
		ClusterService: containerService.Projects.Zones.Clusters,
	}

	return api, nil
}

// getDetails
// as clusterID pass client.Prefix + clusterConfig.Name
func getDetails(ctx context.Context, clusterID string, gcpclient API) (*container.Cluster, error) {
	var cluster *container.Cluster
	var err error
	for i := 0; i < 5; i++ {
		cluster, err = gcpclient.Get(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster details, got: %w", err)
		}
		switch cluster.Status {
		case "RUNNING":
			return cluster, nil
		case "PROVISIONING":
			time.Sleep(60 * time.Second)
		default:
			return nil, fmt.Errorf("failed to get cluster details, cluster state is: %s", cluster.Status)
		}
	}
	return nil, fmt.Errorf("failed to get cluster details, after 5 minutes cluster is not ready, state is: %s", cluster.Status)
}

func NewClient(ctx context.Context, clusterID string, gcpclient API) (*Client, error) {
	details, err := getDetails(ctx, clusterID, gcpclient)
	if err != nil {return nil, fmt.Errorf("failed creating k8s client. got: %w", err)}
	ca, err := base64.StdEncoding.DecodeString(details.MasterAuth.ClusterCaCertificate)
	if err != nil {log.Fatalf("Failed to get cluster ca cert, got: %v", err)}
	config := &rest.Config{
		Host:            details.Endpoint,
		AuthProvider:    &clientcmdapi.AuthProviderConfig{Name: "gcp"},
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {return nil,  fmt.Errorf("failed creating k8s client, got: %w", err)}
	return &Client{K8sclient: clientset}, nil
}
