package installer

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	"log"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
)

type Cleaner struct {
	storageClient  *storage.Client
	clustersClient *cluster.Client
	iamClient      *serviceaccount.Client
	crmClient      *roles.Client
	config         *config.Config
}

func (c *Cleaner) WithClients(storage *storage.Client, cluster *cluster.Client, iam *serviceaccount.Client, crm *roles.Client) *Cleaner {
	if storage == nil || cluster == nil || iam == nil {
		log.Fatalf("failed set clients, passed client can not be nil")
		return nil
	}
	c.storageClient = storage
	c.clustersClient = cluster
	c.iamClient = iam
	c.crmClient = crm
	return c
}

func (c *Cleaner) WithConfig(config installerconfig.Config) *Cleaner {
	c.config = config
	return c
}

func (c *Cleaner) CleanAll(ctx context.Context) error {
	errorslist := make([]string, 0)
	var clusterName string
	for _, v := range c.config.Clusters {
		clusterName = AddPrefix(c.config, v.Name)
		clusterName = TrimName(clusterName)
		err := c.clustersClient.Delete(ctx, clusterName, v.Location)
		if err != nil {
			err = logError(err, "cluster", clusterName)
			errorslist = append(errorslist, err.Error())
		}
	}
	for _, v := range c.config.Buckets {
		var bucketName string
		bucketName = AddPrefix(c.config, v.Name)
		bucketName = TrimName(bucketName)
		err := c.storageClient.DeleteBucket(ctx, bucketName)
		if err != nil {
			err = logError(err, "bucket", bucketName)
			errorslist = append(errorslist, err.Error())
		}
	}
	for _, v := range c.config.ServiceAccounts {
		//TODO: Adding prefix, creating fqdn and adding resource dependent name string should be done within called methods not in calling package. Move such code to called methods for all packages.
		//
		var name string
		name = FormatName(c.config, v.Name)
		_, err := c.crmClient.RemoveSaRole(name, v.Roles, c.config.Project, nil)
		if err != nil {
			log.Fatalf("Failed remove %s from roles, got: %v", name, err)
		}
		safqdn := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, c.config.Project)
		saname := fmt.Sprintf("projects/-/serviceAccounts/%s", safqdn)
		_, err = c.iamClient.Delete(saname)
		if err != nil {
			err = logError(err, "serviceaccount", saname)
			errorslist = append(errorslist, err.Error())
		}
	}
	if len(errorslist) > 0 {
		return fmt.Errorf("failed remove some resources, got: %v", errorslist)
	} else {
		return nil
	}
}

func logError(err error, resourceType string, name string) error {
	message := fmt.Sprintf("failed remove %s %s, got: %v", resourceType, name, err.Error())
	log.Printf(message)
	return fmt.Errorf(message)
}
