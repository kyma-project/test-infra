package release

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// StorageAPI exposes a function to read objects from Google Storage buckets
type StorageAPI interface {
	ReadBucketObject(ctx context.Context, fileName string) (io.ReadCloser, int64, error)
}

// storageAPIWrapper implements reading objects from Google Storage buckets
type storageAPIWrapper struct {
	storageClient *storage.Client
	bucketName    string
}

// NewStorageAPI returns implementation of storageAPI
func NewStorageAPI(ctx context.Context, bucketName string) (StorageAPI, error) {

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &storageAPIWrapper{
		storageClient: storageClient,
		bucketName:    bucketName,
	}, nil
}

// ReadBucketObject downloads and saves in temporary directory a file specified by fileName from a given bucket
func (saw *storageAPIWrapper) ReadBucketObject(ctx context.Context, fileName string) (io.ReadCloser, int64, error) {
	common.Shout("Reading %s file from %s bucket", fileName, saw.bucketName)

	obj := saw.storageClient.Bucket(saw.bucketName).Object(fileName)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, 0, err
	}

	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, 0, err
	}

	return rc, attrs.Size, nil
}
