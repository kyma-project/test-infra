package githubrelease

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// StorageAPIWrapper abstracts Google Storage API
type StorageAPIWrapper struct {
	Context       context.Context
	StorageClient *storage.Client
	BucketName    string
	FolderName    string
}

// ReadBucketObject downloads and saves in temporary directory a file specified by fileName from a given bucket
func (saw *StorageAPIWrapper) ReadBucketObject(fileName string) ([]byte, error) {
	common.Shout("Reading %s/%s file from %s bucket", saw.FolderName, fileName, saw.BucketName)

	rc, err := saw.StorageClient.Bucket(saw.BucketName).Object(saw.FolderName + "/" + fileName).NewReader(saw.Context)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
}
