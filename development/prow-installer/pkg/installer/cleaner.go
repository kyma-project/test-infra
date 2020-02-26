package installer

import (
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/cluster"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"

	installerconfig "github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
)

type Cleaner struct {
	storageClient  storage.Client
	clustersClient cluster.Client
	iamClient      serviceaccount.Client
	config         installerconfig.Config
}

func (c *Cleaner) CleanAll() error {

}
