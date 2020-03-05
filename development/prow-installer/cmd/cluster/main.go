package main

import (
	"context"
	"flag"
	"os"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"

	log "github.com/sirupsen/logrus"
)

var (
	projectID = flag.String("proj", "", "ProjectID of the GCP project [Required]")
	zoneID    = flag.String("zone", "global", "GCP zone for the cluster to be created [Required]")
	prefix    = flag.String("prefix", "", "Prefix of a cluster [Optional]")
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

	// cluster service define
	clusterService, err := cluster.NewService(ctx, *projectID)
	if err != nil {
		log.Fatalf("An error occurred during cluster service configuration: %v", err)
	}

	_, err = cluster.New(*projectID, *prefix, clusterService)
	if err != nil {
		log.Fatalf("An error occurred during cluster client configuration: %v", err)
	}

	labels := make(map[string]string)
	labels["created-for"] = "testing"
}
