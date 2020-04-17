package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=ObjectAttrs -output=automock -outpkg=automock -case=underscore

// ObjectAttrs a GCS bucket object metadata
type ObjectAttrs interface {
	Name() string
	Bucket() string
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=Query -output=automock -outpkg=automock -case=underscore

// Query a bucket object filter query
type Query interface {
	Delimiter() string
	Prefix() string
	Versions() bool
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=ObjectIterator -output=automock -outpkg=automock -case=underscore

// ObjectIterator iterate over bucket object metadata
type ObjectIterator interface {
	Next() (ObjectAttrs, error)
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=ObjectHandle -output=automock -outpkg=automock -case=underscore

// ObjectHandle allows to operate on GCS bucket object
type ObjectHandle interface {
	Delete(ctx context.Context) error
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=Client -output=automock -outpkg=automock -case=underscore

// Client provides interaction with GCS
type Client interface {
	Bucket(bucketName string) BucketHandle
	Buckets(ctx context.Context, projectID string) BucketIterator
	Close() error
}

// NewClient creates a new client to interact with GCP buckets
func NewClient(ctx context.Context) (Client, error) {
	storageClient, err := storage.NewClient(ctx)
	client := client{
		client: storageClient,
	}
	return client, err
}

type client struct {
	client *storage.Client
}

func (r client) Close() error {
	return r.client.Close()
}

func (r client) Bucket(bucketName string) BucketHandle {
	return bucketHandle{
		bucketHandle: r.client.Bucket(bucketName),
	}
}

func (r client) Buckets(ctx context.Context, projectID string) BucketIterator {
	storageBucketIterator := r.client.Buckets(ctx, projectID)
	return NewBucketIterator(storageBucketIterator)
}
