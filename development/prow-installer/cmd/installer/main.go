package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/installer"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/k8s"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"time"
)

var (
	configPath      = flag.String("config", "", "Config file path [Required]")
	credentialsFile = flag.String("credentials-file", "", "Google Application Credentials file path [Required]")
	remove          = flag.Bool("remove", false, "When set, installer will remove resources defined in config. Default false. [Optional]")
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

	// this variable must be set
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", *credentialsFile); err != nil {
		log.Fatalf("Could not set GOOGLE_APPLICATION_CREDENTIALS env variable.")
	}

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

	//TODO: refactor GKE operating packages to use os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") in order to get credentials.
	clusterClient, err := cluster.NewClient(ctx, *clusterConfig, *credentialsFile)
	if err != nil {
		log.Fatalf("An error occurred during cluster client configuration: %v", err)
	}

	storageClient, err := storage.NewClient(ctx, *storageConfig, *credentialsFile)
	if err != nil {
		log.Fatalf("An error occurred during storage client configuration: %v", err)
	}

	iamService, err := serviceaccount.NewService(*credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create IAM service %v", err)
	}
	iamClient := serviceaccount.NewClient(iamService)

	crmService, err := roles.NewService(*credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create CRM service %v", err)
	}
	crmClient, err := roles.New(crmService)

	gkeClient, err := cluster.NewGKEClient(ctx, readConfig.Project)
	if err != nil {
		log.Fatalf("failed get gke client, got: %v", err)
	}

	prowConfig := installer.NewProwConfig()

	if *remove {
		cleaner := installer.Cleaner{}
		cleaner.WithClients(storageClient, clusterClient, iamClient, crmClient).WithConfig(readConfig)
		err := cleaner.CleanAll(ctx)
		if err != nil {
			log.Fatalf("cleaning all resources failed, got: %v", err)
		}
		os.Exit(0)
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

	for _, bucket := range readConfig.Buckets {
		bucketname := installer.AddPrefix(readConfig, bucket.Name)
		bucketname = installer.TrimName(bucketname)
		if bucket.Name == readConfig.GCSlogBucket {
			_ = prowConfig.WithGCSprowBucket(bucketname)
		}
		if err := storageClient.CreateBucket(ctx, bucketname, bucket.Location); err != nil {
			log.Fatalf("Failed to create bucket: %s, %s", bucket, err)
		}
	}

	for i, serviceAccount := range readConfig.ServiceAccounts {
		// TODO implement handling error when SA already exists in GCP
		saname := installer.FormatName(readConfig, serviceAccount.Name)
		if serviceAccount.Name == readConfig.GCSserviceAccount {
			_ = prowConfig.WithGCScredentials(saname)
		}
		readConfig.ServiceAccounts[i].Name = saname
		sa, err := iamClient.CreateSA(saname, readConfig.Project)
		if err != nil {
			log.Errorf("Error creating Service Account %v", err)
		} else {
			key, err := iamClient.CreateSAKey(sa.Email)
			if err != nil {
				log.Errorf("failed create serviceaccount %s key, got: %w", sa.Name, err)
			}
			readConfig.ServiceAccounts[i].Key = key
			if len(serviceAccount.Roles) > 0 {
				_, err = crmClient.AddSAtoRole(saname, serviceAccount.Roles, readConfig.Project, nil)
				if err != nil {
					log.Errorf("Failed assign sa %s to roles, got: %v", saname, err)
				}
			}
		}
	}

	for k, v := range readConfig.Clusters {
		clusterID := installer.AddPrefix(readConfig, v.Name)
		clusterID = installer.TrimName(clusterID)
		k8sclient, kubectlWrapper, err := k8s.NewClient(ctx, clusterID, v.Location, gkeClient)
		if err != nil {
			log.Fatalf("failed get k8s client, got: %v", err)
		}
		v.K8sClient = k8sclient
		v.KubectlWrapper = kubectlWrapper
		populator := k8s.NewPopulator(k8sclient)
		v.Populator = populator
		readConfig.Clusters[k] = v
		if err := populator.PopulateSecrets(metav1.NamespaceDefault, readConfig.GenericSecrets, readConfig.ServiceAccounts); err != nil {
			log.Fatalf("failed populate secrets, got %v", err)
		}
	}
	// TODO: Below lines should be uncommented when igress IP will be available for prow installer.
	//ingress, err := readConfig.Clusters["prow"].K8sClient.NetworkingClient.Ingresses(metav1.NamespaceDefault).Get("tls-ing", metav1.GetOptions{})
	//if err != nil {
	//	log.Fatalf("Failed getting ingress IP.")
	//}
	fmt.Printf("%+v", prowConfig)
	//clusterIP := ingress.Status.LoadBalancer.Ingress[0].IP
	//_ = prowConfig.WithClusterIP(clusterIP)
}
