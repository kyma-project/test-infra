package main

import (
	"context"
	"flag"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/accessmanager"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/secrets"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/installer"
)

var (
	config          = flag.String("config", "", "Config file path [Required]")
	credentialsfile = flag.String("credentials-file", "", "Google Application Credentials file path [Required]")
)

func main() {
	flag.Parse()

	if *config == "" {
		log.Fatalf("Missing required argument : -config")
	}

	if *credentialsfile == "" {
		log.Fatalf("Missing required argument : -credentials-file")
	}

	var InstallerConfig installer.InstallerConfig
	InstallerConfig.ReadConfig(*config)
	ctx := context.Background()

	storageConfig := &storage.Option{
		ProjectID:      InstallerConfig.Project,
		LocationID:     InstallerConfig.Location,
		Prefix:         InstallerConfig.Prefix,
		ServiceAccount: *credentialsfile,
	}
	clusterConfig := &cluster.Option{
		ProjectID:      InstallerConfig.Project,
		ZoneID:         InstallerConfig.Zone,
		ServiceAccount: *credentialsfile,
	}
	secretsConfig := &secrets.Option{
		ProjectID:      InstallerConfig.Project,
		LocationID:     InstallerConfig.Location,
		KmsRing:        InstallerConfig.KeyringName,
		KmsKey:         InstallerConfig.EncryptionKeyName,
		ServiceAccount: *credentialsfile,
	}

	clusterClient, err := cluster.NewClient(ctx, *clusterConfig, *credentialsfile)
	if err != nil {
		log.Fatalf("An error occurred during cluster client configuration: %v", err)
	}
	if err := clusterClient.Create(ctx, InstallerConfig.ClusterName, nil, 3, false); err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	storageClient, err := storage.NewClient(ctx, *storageConfig, *credentialsfile)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}
	if err := storageClient.CreateBucket(ctx, InstallerConfig.BucketName); err != nil {
		log.Fatalf("Failed to create bucket: %s, %s", InstallerConfig.BucketName, err)
	}

	AccessManager := accessmanager.NewAccessManager(*credentialsfile)
	for _, account := range InstallerConfig.ServiceAccounts {
		_ = AccessManager.IAM.CreateSAAccount(account.Name, InstallerConfig.Project)
	}
	AccessManager.Projects.GetProjectPolicy(InstallerConfig.Project)
	log.Printf("%+v", AccessManager.Projects.Projects[InstallerConfig.Project].Policy)
	//AccessManager.Projects.AssignRoles(InstallerConfig.Project, InstallerConfig.ServiceAccounts)

	secretsClient, err := secrets.NewClinet(ctx, *secretsConfig, *credentialsfile)
	if err != nil {
		log.Fatalf("Failed to create secrets client: %v", err)
	}

	if data, err := secretsClient.Encrypt(ctx, []byte("super secret string")); err != nil {
		log.Errorf("Failed to encrypt: %v", err)
	} else if err := storageClient.Write(ctx, data, InstallerConfig.BucketName, "mySecret.encrypted"); err != nil {
		log.Errorf("Failed to write to bucket %s: %v", InstallerConfig.BucketName, err)
	}
}
