package gcscleaner

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"google.golang.org/api/iterator"
)

type bucket struct {
	attrs *storage.BucketAttrs
}

type client struct {
	stiface.Client
	buckets map[string]*bucket
}

type bucketHandle struct {
	stiface.BucketHandle
	c    *client
	name string
}

func (b bucketHandle) Delete(ctx context.Context) error {
	delete(b.c.buckets, b.name)
	return nil
}

func NewFakeClient(data []string) stiface.Client {
	buckets := map[string]*bucket{}
	for _, name := range data {
		a := storage.BucketAttrs{Name: name}
		buckets[name] = &bucket{attrs: &a}
	}
	return &client{buckets: buckets}
}

func (c *client) Bucket(name string) stiface.BucketHandle {
	return bucketHandle{c: c, name: name}
}

type bucketIterator struct {
	buckets []*bucket
	stiface.BucketIterator
	index int
}

func (b *bucketIterator) Next() (*storage.BucketAttrs, error) {
	if len(b.buckets) < b.index+1 {
		return nil, iterator.Done
	}
	attrs := b.buckets[b.index].attrs
	b.index++
	return attrs, nil
}

func (c *client) Buckets(ctx context.Context, projectID string) stiface.BucketIterator {
	var buckets []*bucket
	for _, value := range c.buckets {
		buckets = append(buckets, value)
	}
	return &bucketIterator{buckets: buckets}
}

func GetBucketNames(bucketIterator stiface.BucketIterator) ([]string, error) {
	var buckets []string
	for {
		attr, err := bucketIterator.Next()
		if err == iterator.Done {
			return buckets, nil
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, attr.Name)
	}
}
