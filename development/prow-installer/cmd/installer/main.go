package main

import (
	"context"
	"flag"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/installer"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"
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

	var Config installer.Config
	if err := Config.ReadConfig(*config); err != nil {
		log.Fatalf("Error reading config file %v", err)
	}
	//Config.Labels["created-on"] = time.Now()
	ctx := context.Background()

	storageConfig := &storage.Option{
		ProjectID:      Config.Project,
		LocationID:     Config.Location,
		Prefix:         Config.Prefix,
		ServiceAccount: *credentialsfile,
	}
	clusterConfig := &cluster.Option{
		ProjectID:      Config.Project,
		ZoneID:         Config.Zone,
		ServiceAccount: *credentialsfile,
	}
	//secretsConfig := &secrets.Option{
	//	ProjectID:      Config.Project,
	//	LocationID:     Config.Location,
	//	KmsRing:        Config.KeyringName,
	//	KmsKey:         Config.EncryptionKeyName,
	//	ServiceAccount: *credentialsfile,
	//}

	clusterClient, err := cluster.NewClient(ctx, *clusterConfig, *credentialsfile)
	if err != nil {
		log.Fatalf("An error occurred during cluster client configuration: %v", err)
	}
	if err := clusterClient.Create(ctx, Config.ClusterName, Config.Labels, 3, false); err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	storageClient, err := storage.NewClient(ctx, *storageConfig, *credentialsfile)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}
	if err := storageClient.CreateBucket(ctx, Config.BucketName); err != nil {
		log.Fatalf("Failed to create bucket: %s, %s", Config.BucketName, err)
	}

	//AccessManager := accessmanager.NewAccessManager(*credentialsfile)
	//for _, account := range Config.ServiceAccounts {
	//	_ = AccessManager.IAM.CreateSAAccount(account.Name, Config.Project)
	//}
	//AccessManager.Projects.GetProjectPolicy(Config.Project)
	//log.Printf("%+v", AccessManager.Projects.Projects[Config.Project].Policy)
	////AccessManager.Projects.AssignRoles(Config.Project, Config.ServiceAccounts)

	//secretsClient, err := secrets.NewClinet(ctx, *secretsConfig, *credentialsfile)
	//if err != nil {
	//	log.Fatalf("Failed to create secrets client: %v", err)
	//}
	//
	//if data, err := secretsClient.Encrypt(ctx, []byte("super secret string")); err != nil {
	//	log.Errorf("Failed to encrypt: %v", err)
	//} else if err := storageClient.Write(ctx, data, Config.BucketName, "mySecret.encrypted"); err != nil {
	//	log.Errorf("Failed to write to bucket %s: %v", Config.BucketName, err)
	//}
}
