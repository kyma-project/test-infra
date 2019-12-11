package main

import (
	"context"
	"flag"
	"os"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"

	log "github.com/sirupsen/logrus"

	container "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// http://blog.ralch.com/categories/design-patterns/

var (
	projectID  = flag.String("proj", "", "ProjectID of the GCP project [Required]")
	locationID = flag.String("loc", "global", "Location of the keyring used for encryption/decryption [Optional]")
)

func main() {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatalf("Requires the environment variable GOOGLE_APPLICATION_CREDENTIALS to be set to a GCP service account file.")
	}

	flag.Parse()
	if *projectID == "" {
		log.Fatalf("Missing required argument : -proj")
	}
	ctx := context.Background()

	containerService, err := container.NewService(ctx, option.WithServiceAccountFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		log.Fatalf("Couldn't create service handle for GCP: %w", err)
	}
	clusterService := containerService.Projects.Locations.Clusters

	wrappedAPI := &cluster.APIWrapper{
		ClusterService: clusterService,
	}

	gkeClient, err := cluster.New(cluster.Option{ProjectID: *projectID, LocationID: *locationID, ServiceAccount: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")}, wrappedAPI)
	if err != nil {
		log.Errorf("Could not create GKE Client: %v", err)
		os.Exit(1)
	}

	gkeClient.Create(ctx, "hello")
}
