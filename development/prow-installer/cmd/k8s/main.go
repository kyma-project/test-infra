package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
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

	containerService, err := container.NewService(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("failed creating gke client, got: %v", err)
	}

	api := &cluster.APIWrapper{
		ProjectID:      *projectID,
		ZoneID:         *zoneID,
		ClusterService: containerService.Projects.Zones.Clusters,
	}

	k8sclient, err := k8s.NewClient(ctx, *clusterID, api)
	if err != nil {
		log.Fatalf("failed create k8s client, got: %v", err)
	}
	secretlist, err := k8sclient.K8sclient.CoreV1().Secrets(metav1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("failed list secrets, got: %v", err)
	}
	fmt.Print(secretlist.Items)
}
