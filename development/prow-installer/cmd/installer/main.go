package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	configPath      = flag.String("config", "", "Config file path [Required]")
	credentialsFile = flag.String("credentials-file", "", "Google Application Credentials file path [Required]")
)

func main() {
	flag.Parse()

	if *configPath == "" {
		log.Fatalf("Missing required argument : -configPath")
	}

	if *credentialsFile == "" {
		log.Fatalf("Missing required argument : -credentials-file")
	}

	readConfig, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error reading configPath file %v", err)
	}
	readConfig.Labels["created-at"] = fmt.Sprintf("%v", time.Now().Unix()) // time of cluster creation

	ctx := context.Background()

	storageConfig := &storage.Option{
		ProjectID:      readConfig.Project,
		LocationID:     readConfig.Location,
		Prefix:         readConfig.Prefix,
		ServiceAccount: *credentialsFile,
	}
	clusterConfig := &cluster.Option{
		Prefix:         readConfig.Prefix,
		ProjectID:      readConfig.Project,
		ZoneID:         readConfig.Zone,
		ServiceAccount: *credentialsFile,
	}

	clusterClient, err := cluster.NewClient(ctx, *clusterConfig, *credentialsFile)
	if err != nil {
		log.Fatalf("An error occurred during cluster client configuration: %v", err)
	}
	if err := clusterClient.Create(ctx, readConfig.ClusterName, readConfig.Labels, 1, false); err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	storageClient, err := storage.NewClient(ctx, *storageConfig, *credentialsFile)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}
	if err := storageClient.CreateBucket(ctx, readConfig.BucketName); err != nil {
		log.Fatalf("Failed to create bucket: %s, %s", readConfig.BucketName, err)
	}

	iamService, err := serviceaccount.NewService(*credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create IAM service %v", err)
	}
	crmService, err := roles.NewService(*credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create CRM service %v", err)
	}

	iamClient := serviceaccount.NewClient(readConfig.Prefix, &iamService)
	crmClient, err := roles.New(crmService)

	for _, serviceAccount := range readConfig.ServiceAccounts {
		opts := serviceaccount.SAOptions{
			Name:    serviceAccount.Name,
			Roles:   serviceAccount.Roles,
			Project: readConfig.Project,
		}
		// TODO implement handling error when SA already exists in GCP
		if _, err := iamClient.CreateSA(opts); err != nil {
			log.Errorf("Error creating Service Account %v", err)
		} else {
			_, err = crmClient.AddSAtoRole(serviceAccount.Name, serviceAccount.Roles, readConfig.Project, nil)
			if err != nil {
				log.Errorf("Error adding role to a service %v", err)
			}
		}
	}
}
