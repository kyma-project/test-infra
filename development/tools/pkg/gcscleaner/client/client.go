package client

import (
	"cloud.google.com/go/storage"
	"context"
)

//go:generate mockery -name=BucketAttrs -output=automock -outpkg=automock -case=underscore
type BucketAttrs interface {
	Name() string
}

type bucketAttrs struct {
	bucketAttrs *storage.BucketAttrs
}

func (r bucketAttrs) Name() string {
	return r.bucketAttrs.Name
}

//go:generate mockery -name=Query -output=automock -outpkg=automock -case=underscore
type Query interface {
	Delimiter() string
	Prefix() string
	Versions() bool
}

//go:generate mockery -name=ObjectAttrs -output=automock -outpkg=automock -case=underscore
type ObjectAttrs interface {
	Name() string
	Bucket() string
}

//go:generate mockery -name=ObjectIterator -output=automock -outpkg=automock -case=underscore
type ObjectIterator interface {
	Next() (ObjectAttrs, error)
}

//go:generate mockery -name=ObjectHandle -output=automock -outpkg=automock -case=underscore
type ObjectHandle interface {
	Delete(ctx context.Context) error
}

//go:generate mockery -name=BucketHandle -output=automock -outpkg=automock -case=underscore
type BucketHandle interface {
	Object(name string) ObjectHandle
	Objects(ctx context.Context, q Query) ObjectIterator
	Delete(ctx context.Context) (err error)
}

//go:generate mockery -name=BucketIterator -output=automock -outpkg=automock -case=underscore
type BucketIterator interface {
	Next() (BucketAttrs, error)
}

type bucketIterator struct {
	bucketIterator *storage.BucketIterator
}

func (r bucketIterator) Next() (BucketAttrs, error) {
	storageBucketAttrs, err := r.bucketIterator.Next()
	if err != nil {
		return nil, err
	}
	bucketAttrs := bucketAttrs{
		bucketAttrs: storageBucketAttrs,
	}
	return bucketAttrs, nil
}

func NewBucketIterator(iterator *storage.BucketIterator) BucketIterator {
	return bucketIterator{
		bucketIterator: iterator,
	}
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Bucket(bucketName string) BucketHandle
	Buckets(ctx context.Context, projectID string) BucketIterator
	Close() error
}

//go:generate mockery -name=BucketObject -output=automock -outpkg=automock -case=underscore
type BucketObject interface {
	Name() string
	Bucket() string
}

type bucketObject struct {
	name   string
	bucket string
}

func NewBucketObject(bucket string, name string) BucketObject {
	return bucketObject{
		name:   name,
		bucket: bucket,
	}
}

func (r bucketObject) Name() string {
	return r.name
}

func (r bucketObject) Bucket() string {
	return r.bucket
}

func NewClient2(ctx context.Context) (Client, error) {
	storageClient, err := storage.NewClient(ctx)
	client := client{
		client: storageClient,
	}
	return client, err
}

type objectAttrs struct {
	name   string
	bucket string
}

func (r objectAttrs) Name() string {
	return r.name
}

func (r objectAttrs) Bucket() string {
	return r.bucket
}

type objectIterator struct {
	objectIterator *storage.ObjectIterator
}

func (r objectIterator) Next() (ObjectAttrs, error) {
	storageObjectAttrs, err := r.objectIterator.Next()
	if err != nil {
		return nil, err
	}
	return objectAttrs{
		name:   storageObjectAttrs.Name,
		bucket: storageObjectAttrs.Bucket,
	}, err
}

type client struct {
	client *storage.Client
}

func (r client) Close() error {
	return r.client.Close()
}

type bucketHandle struct {
	bucketHandle *storage.BucketHandle
}

func (r bucketHandle) Object(name string) ObjectHandle {
	return r.bucketHandle.Object(name)
}

func (r bucketHandle) Objects(ctx context.Context, q Query) ObjectIterator {
	return objectIterator{
		objectIterator: r.bucketHandle.Objects(ctx, nil),
	}
}

func (r bucketHandle) Delete(ctx context.Context) (err error) {
	return r.bucketHandle.Delete(ctx)
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
