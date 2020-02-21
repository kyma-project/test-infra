package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/kubehelper"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"
	"os"
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
		log.Fatalf("Error reading config file %v", err)
	}
	// define mandatory environment variables
	_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *credentialsFile)

	kubeHelpers := make(map[string]*kubehelper.KubeHelper) // map of k8s cluster helpers to communicate with them

	readConfig.Labels["created-at"] = fmt.Sprintf("%v", time.Now().Unix()) // time of cluster creation

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
		if kubeConfig, err := clusterClient.Create(ctx, clusterToCreate); err != nil {
			log.Fatalf("Failed to create cluster: %v", err)
		} else {
			// TODO parametrize types of cluster for defining them later
			kubeHelpers["default-cluster"] = &kubehelper.KubeHelper{Kubeconfig: kubeConfig}
		}
	}
	// TODO selection of a cluster based on a parameter
	for {
		if err := kubeHelpers["default-cluster"].Apply().File("https://k8s.io/examples/controllers/nginx-deployment.yaml").Do(); err != nil {
			log.Errorf("Can't apply file. %v Retrying in 5 seconds...", err) //probing a cluster until it responds
			time.Sleep(time.Second * 5)
			continue
		}
		break
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

	for _, serviceAccount := range readConfig.ServiceAccounts {
		// TODO implement handling error when SA already exists in GCP
		if _, err := iamClient.CreateSA(serviceAccount.Name, readConfig.Project); err != nil {
			log.Errorf("Error creating Service Account %v", err)
		} else {
			//log.Println(iamClient.CreateSAKey(sa.Email))
			_, err = crmClient.AddSAtoRole(serviceAccount.Name, serviceAccount.Roles, readConfig.Project, nil)
		}
	}
}
