package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"

	log "github.com/sirupsen/logrus"

	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

var (
	projectID = flag.String("proj", "", "ProjectID of the GCP project [Required]")
	zoneID    = flag.String("zone", "global", "GCP zone for the cluster to be created [Required]")
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
	ctx := context.Background()

	containerService, err := container.NewService(ctx, option.WithServiceAccountFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Couldn't create service handle for GCP: %w", err))
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

	labels := make(map[string]string)
	labels["created-for"] = "testing"

	// err = gkeClient.Create(ctx, "daniel-test-cluster", labels, 1, true)
	// if err != nil {
	// 	log.Fatalf("Couldn't create cluster %w", err)
	// }
	err = gkeClient.Delete(ctx, "", "europe-west-3-c")
	if err != nil {
		log.Fatalf("%v", fmt.Errorf("Couldn't delete cluster %w", err))
	}
}
