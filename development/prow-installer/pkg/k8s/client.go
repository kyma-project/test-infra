package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/kubectl"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/kubernetes"
	networking "k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"path/filepath"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type API interface {
	Get(ctx context.Context, clusterID string, zoneID string) (*container.Cluster, error)
}

type K8sClient struct {
	Clientset        *kubernetes.Clientset
	NetworkingClient *networking.NetworkingV1beta1Client
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

func NewClient(ctx context.Context, clusterID string, zoneID string, gcpclient API) (*K8sClient, *kubectl.Wrapper, error) {
	details, err := getDetails(ctx, clusterID, zoneID, gcpclient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating k8s client. got: %w", err)
	}

	kubeconfigPath, err := generateKubeconfig(details.Endpoint, details.MasterAuth.ClusterCaCertificate, clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating kubeconfig file %w", err)
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
	networkingClient, err := networking.NewForConfig(config)
	k8sClient := &K8sClient{
		Clientset:        clientset,
		NetworkingClient: networkingClient,
	}
	kubectlWrapper := &kubectl.Wrapper{Kubeconfig: kubeconfigPath}

	if err != nil {
		return nil, nil, fmt.Errorf("failed creating k8s client, got: %w", err)
	}
	return k8sClient, kubectlWrapper, nil
}

// generateKubeconfig generates kubeconfig based on credentials provided in arguments
// the function returns path to the config file.
// it's needed to have GOOGLE_CREDENTIALS_APPLICATION env variable set
func generateKubeconfig(endpoint, cadata, name string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("endpoint cannot be empty")
	}
	if cadata == "" {
		return "", fmt.Errorf("cadata cannot be empty")
	}
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory %w", err)
	}

	path := filepath.FromSlash(cwd + "/.kube/")
	file := name + "_config"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Debugf("%s does not exist. Creating directory...", path)
		if err := os.MkdirAll(path, 0700); err != nil {
			return "", fmt.Errorf("unexpected error during folder creation %w", err)
		}
	}
	log.Debugf("Creating path %s", path+file)
	f, err := os.Create(path + file)
	if err != nil {
		return "", fmt.Errorf("error creating kubeconfig file %w", err)
	}
	defer f.Close()

	log.Debugf("Generating GCP Kubeconfig file with credentials for host: %s, name: %s", endpoint, name)
	kubeconfigTemplate := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: gke-cluster
users:
- name: gke-user
  user:
    auth-provider:
      name: gcp
contexts:
- context:
    cluster: gke-cluster
    user: gke-user
  name: gke-cluster
current-context: gke-cluster`, cadata, endpoint)
	if _, err = f.WriteString(kubeconfigTemplate); err != nil {
		return "", fmt.Errorf("error writing to kubeconfig file %w", err)
	}
	return path, nil
}
