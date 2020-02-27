package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"log"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type API interface {
	Get(ctx context.Context, clusterID string, zoneID string) (*container.Cluster, error)
}

// getDetails
// as clusterID pass client.Prefix + clusterConfig.Name
func getDetails(ctx context.Context, clusterID string, zoneID string, gcpclient API) (*container.Cluster, error) {
	var gkecluster *container.Cluster
	var err error
	for i := 0; i < 5; i++ {
		gkecluster, err = gcpclient.Get(ctx, clusterID, zoneID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster details, got: %w", err)
		}
		switch gkecluster.Status {
		case "RUNNING":
			log.Printf("Cluster %s provisioned.", clusterID)
			return gkecluster, nil
		case "PROVISIONING":
			log.Printf("Cluster %s is still in PROVISIONING state.", clusterID)
			time.Sleep(60 * time.Second)
		default:
			return nil, fmt.Errorf("failed to get cluster details, cluster state is: %s", gkecluster.Status)
		}
	}
	return nil, fmt.Errorf("failed to get cluster details, after 5 minutes cluster is not ready, state is: %s", gkecluster.Status)
}

func NewClient(ctx context.Context, clusterID string, zoneID string, gcpclient API) (*kubernetes.Clientset, error) {
	details, err := getDetails(ctx, clusterID, zoneID, gcpclient)
	if err != nil {
		return nil, fmt.Errorf("failed creating k8s client. got: %w", err)
	}
	ca, err := base64.StdEncoding.DecodeString(details.MasterAuth.ClusterCaCertificate)
	if err != nil {
		log.Fatalf("Failed to get cluster ca cert, got: %v", err)
	}
	config := &rest.Config{
		Host:         details.Endpoint,
		AuthProvider: &clientcmdapi.AuthProviderConfig{Name: "gcp"},
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed creating k8s client, got: %w", err)
	}
	return clientset, nil
}
