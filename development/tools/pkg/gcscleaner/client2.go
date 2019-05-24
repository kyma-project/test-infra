package gcscleaner

import (
	"cloud.google.com/go/storage"
	"context"
	"sync"
)

//TODO move all client related interfaces here

func (r Cleaner2) deleteAllObjects(
	ctx CancelableContext,
	bucketName string,
	errChan chan error) {
	defer close(errChan)

	bucketObjectChan := make(chan BucketObject)
	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		r.iterateBucketObjectNames(ctx, bucketName, bucketObjectChan, errChan)
	}()

	for i := 0; i < r.cfg.BucketObjectWorkersNumber; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			r.deleteBucketObjects(ctx, bucketObjectChan, errChan)
		}()
	}
	waitGroup.Wait()
}

func NewClient2(ctx context.Context) (Client, error) {
	storageClient, err := storage.NewClient(ctx)
	client := client2{
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

type client2 struct {
	client *storage.Client
}

func (r client2) Close() error {
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

func (r client2) Bucket(bucketName string) BucketHandle {
	return bucketHandle{
		bucketHandle: r.client.Bucket(bucketName),
	}
}

func (r client2) Buckets(ctx context.Context, projectID string) BucketIterator {
	storageBucketIterator := r.client.Buckets(ctx, projectID)
	return bucketIterator{
		bucketIterator: storageBucketIterator,
	}
}
