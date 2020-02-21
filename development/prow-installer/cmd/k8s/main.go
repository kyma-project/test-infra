package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

var (
	projectID = flag.String("proj", "", "ProjectID of the GCP project [Required]")
	zoneID    = flag.String("zone", "global", "GCP zone for the cluster to be created [Required]")
	clusterID = flag.String("cluster", "", "GKE cluster ID [Required]")
)


func main() {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatalf("Requires the environment variable GOOGLE_APPLICATION_CREDENTIALS to be set to a GCP service account file.")
	}

	flag.Parse()
	if *projectID == "" {
		log.Fatalf("Missing required argument : -proj")
	}
	if *zoneID == "" {
		log.Fatalf("Missing required argument : -zone")
	}
	if *clusterID == "" {
		log.Fatalf("Missing required argument : -cluster")
	}
	ctx := context.Background()

	containerService, err := container.NewService(ctx, option.WithServiceAccountFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("Couldn't create service handle for GCP: %w", err)
	}
	clusterService := containerService.Projects.Zones.Clusters

	wrappedAPI := &cluster.APIWrapper{
		ProjectID:      *projectID,
		ZoneID:         *zoneID,
		ClusterService: clusterService,
	}

	clientOpts := cluster.Option{}
	clientOpts = clientOpts.WithProjectID(*projectID).WithServiceAccount(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	gkeClient, err := cluster.New(clientOpts, wrappedAPI)
	if err != nil {
		log.Errorf("Could not create GKE Client: %v", err)
		os.Exit(1)
	}

	k8sclient, err := k8s.NewClient(ctx, *clusterID, gkeClient)
	if err != nil {
		log.Fatalf("failed create k8s client, got: %w", err)
	}
	secretlist, err := k8sclient.K8sclient.CoreV1().Secrets(corev1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {log.Fatalf("failed list secrets, got: %w", err)}
	fmt.Println(secretlist.Items)
}