package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
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

	gkeClient, err := cluster.NewGKEClient(ctx, *projectID)
	if err != nil {
		log.Fatalf("failed get gke client, got: %v", err)
	}

	k8sclient, _, err := k8s.NewClient(ctx, *clusterID, *zoneID, gkeClient)

	if err != nil {
		log.Fatalf("failed get k8s client, got: %v", err)
	}

	//TODO: Implement logic which will load provided secret in to provided cluster.
	secretlist, err := k8sclient.Clientset.CoreV1().Secrets(metav1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("failed list secrets, got: %v", err)
	}
	fmt.Print(secretlist.Items)
}
