package storage

import (
	"context"

	"cloud.google.com/go/storage"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=BucketHandle -output=automock -outpkg=automock -case=underscore

// BucketHandle allows to operate on
type BucketHandle interface {
	Object(name string) ObjectHandle
	Objects(ctx context.Context, q Query) ObjectIterator
	Delete(ctx context.Context) (err error)
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

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=BucketAttrs -output=automock -outpkg=automock -case=underscore

// BucketAttrs a GCS bucket metadata
type BucketAttrs interface {
	Name() string
}

type bucketAttrs struct {
	bucketAttrs *storage.BucketAttrs
}

func (r bucketAttrs) Name() string {
	return r.bucketAttrs.Name
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=BucketIterator -output=automock -outpkg=automock -case=underscore

// BucketIterator iterator over bucket metadata
type BucketIterator interface {
	Next() (BucketAttrs, error)
}

type bucketIterator struct {
	bucketIterator *storage.BucketIterator
}

// Next result
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

// NewBucketIterator creates new BucketIterator
func NewBucketIterator(iterator *storage.BucketIterator) BucketIterator {
	return bucketIterator{
		bucketIterator: iterator,
	}
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=BucketObject -output=automock -outpkg=automock -case=underscore

// BucketObject identifies GCS object to be deleted
type BucketObject interface {
	Name() string
	Bucket() string
}

type bucketObject struct {
	name   string
	bucket string
}

// Name a bucket object name
func (r bucketObject) Name() string {
	return r.name
}

// Bucket a bucket name
func (r bucketObject) Bucket() string {
	return r.bucket
}

// NewBucketObject creates new bucket object
func NewBucketObject(bucket string, name string) BucketObject {
	return bucketObject{
		name:   name,
		bucket: bucket,
	}
}

type objectAttrs struct {
	name   string
	bucket string
}

// Name a bucket object name
func (r objectAttrs) Name() string {
	return r.name
}

// Bucket a bucket name
func (r objectAttrs) Bucket() string {
	return r.bucket
}

type objectIterator struct {
	objectIterator *storage.ObjectIterator
}

// Next result
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
