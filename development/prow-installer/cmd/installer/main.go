package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

var (
	configPath      = flag.String("config", "", "Config file path [Required]")
	credentialsFile = flag.String("credentials-file", "", "Google Application Credentials file path [Required]")
)

func main() {
	flag.Parse()

	if *configPath == "" {
		log.Fatalf("Missing required argument : -config")
	}

	if *credentialsFile == "" {
		log.Fatalf("Missing required argument : -credentials-file")
	}

	readConfig, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error reading configPath file %v", err)
	}

	//TODO: this should be moved to cluster package in to Create method.
	readConfig.Labels["created-at"] = fmt.Sprintf("%v", time.Now().Unix()) // time of cluster creation
	//TODO: Set label[created-by] to value of client_email from google application credentials file.

	ctx := context.Background()

	storageConfig := &storage.Option{
		ProjectID:      readConfig.Project,
		Prefix:         readConfig.Prefix,
		ServiceAccount: *credentialsFile,
	}

	clusterConfig := &cluster.Option{
		Prefix:         readConfig.Prefix,
		ProjectID:      readConfig.Project,
		ServiceAccount: *credentialsFile,
	}

	//TODO: we need to refactor cluster package or installer package to not pass credentialsfiles data twice in a call to NewCLient method.
	clusterClient, err := cluster.NewClient(ctx, *clusterConfig, *credentialsFile)
	if err != nil {
		log.Fatalf("An error occurred during cluster client configuration: %v", err)
	}
	for _, clusterToCreate := range readConfig.Clusters {
		if clusterToCreate.Labels == nil {
			clusterToCreate.Labels = make(map[string]string)
		}
		for k, v := range readConfig.Labels {
			clusterToCreate.Labels[k] = v
		}
		if err := clusterClient.Create(ctx, clusterToCreate); err != nil {
			log.Fatalf("Failed to create cluster: %v", err)
		}
	}

	storageClient, err := storage.NewClient(ctx, *storageConfig, *credentialsFile)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}
	for _, bucket := range readConfig.Buckets {
		if err := storageClient.CreateBucket(ctx, bucket.Name, bucket.Location); err != nil {
			log.Fatalf("Failed to create bucket: %s, %s", bucket, err)
		}
	}

	iamService, err := serviceaccount.NewService(*credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create IAM service %v", err)
	}
	crmService, err := roles.NewService(*credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create CRM service %v", err)
	}

	iamClient := serviceaccount.NewClient(readConfig.Prefix, iamService)
	crmClient, err := roles.New(crmService)

	for i, serviceAccount := range readConfig.ServiceAccounts {
		// TODO implement handling error when SA already exists in GCP
		if sa, err := iamClient.CreateSA(serviceAccount.Name, readConfig.Project); err != nil {
			log.Errorf("Error creating Service Account %v", err)
		} else {
			key, err := iamClient.CreateSAKey(sa.Email)
			if err != nil {
				log.Errorf("failed create serviceaccount %s key, got: %w", sa.Name, err)
			}
			//log.Println(iamClient.CreateSAKey(sa.Email))
			readConfig.ServiceAccounts[i].Key = key
			_, err = crmClient.AddSAtoRole(serviceAccount.Name, serviceAccount.Roles, readConfig.Project, nil)
			if err != nil {
				log.Errorf("Failed assign sa %s to roles, got: %w", serviceAccount.Name, err)
			}
		}
	}
	gkeClient, err := cluster.NewGKEClient(ctx, readConfig.Project)
	if err != nil {
		log.Fatalf("failed get gke client, got: %v", err)
	}
	var k8sclient *kubernetes.Clientset
	for k, v := range readConfig.Clusters {
		clusterID := fmt.Sprintf("%s-%s", readConfig.Prefix, v.Name)
		k8sclient, err = k8s.NewClient(ctx, clusterID, v.Location, gkeClient)
		if err != nil {
			log.Fatalf("failed get k8s client, got: %v", err)
		}
		v.K8sClient = k8sclient
		populator := k8s.NewPopulator(k8sclient)
		v.Populator = populator
		readConfig.Clusters[k] = v
		err := populator.PopulateSecrets(metav1.NamespaceDefault, readConfig.GenericSecrets, readConfig.ServiceAccounts)
		if err != nil {
			log.Fatalf("failed populate secrets, got %v", err)
		}
	}
}
