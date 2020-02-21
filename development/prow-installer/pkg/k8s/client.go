package k8s


import (
	"context"
	"fmt"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)


// Use GCP client implementaion from prow-installer/cluster package
type API interface {
	Get(ctx context.Context, clusterID string) (*container.Cluster, error)
}

type Client struct {
	K8sclient *kubernetes.Clientset
}

// getDetails
// as clusterID pass client.Prefix + clusterConfig.Name
func getDetails(ctx context.Context, clusterID string, gcpclient API) (*container.Cluster, error) {
	cluster, err := apiclient.Get(ctx, clusterID)
	if err != nil {return nil, fmt.Errorf("failed to get cluster details, got: %w", err)}
	return cluster, nil
}

func NewClient(ctx context.Context, clusterID string, gcpclient API) (*Client, error) {
	details, err := getDetails(ctx, clusterID, gcpclient)
	if err != nil {return nil, fmt.Errorf("failed creating k8s client. got: %w", err)}
	config := &rest.Config{
		Host:            details.Endpoint,
		AuthProvider:    &clientcmdapi.AuthProviderConfig{Name: "gcp"},
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(details.MasterAuth.ClusterCaCertificate),
		},
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {return nil,  fmt.Errorf("failed creating k8s client, got: %w", err)}
	return &Client{k8sclient: clientset}, nil
}
