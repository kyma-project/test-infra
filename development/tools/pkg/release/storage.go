package release

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// StorageAPI exposes a function to read objects from Google Storage buckets
type StorageAPI interface {
	ReadBucketObject(ctx context.Context, fileName string) ([]byte, error)
}

// storageAPIWrapper implements reading objects from Google Storage buckets
type storageAPIWrapper struct {
	storageClient *storage.Client
	bucketName    string
	folderName    string
}

// NewStorageAPI returns implementation of storageAPI
func NewStorageAPI(ctx context.Context, bucketName, folderName string) (StorageAPI, error) {

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &storageAPIWrapper{
		storageClient: storageClient,
		bucketName:    bucketName,
		folderName:    folderName,
	}, nil
}

// ReadBucketObject downloads and saves in temporary directory a file specified by fileName from a given bucket
func (saw *storageAPIWrapper) ReadBucketObject(ctx context.Context, fileName string) ([]byte, error) {
	common.Shout("Reading %s/%s file from %s bucket", saw.folderName, fileName, saw.bucketName)

	rc, err := saw.storageClient.Bucket(saw.bucketName).Object(saw.folderName + "/" + fileName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
}
