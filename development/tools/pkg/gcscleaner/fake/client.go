package fake

import (
	"google.golang.org/api/iterator"
)

type Client struct {
	Buckets      func() []string
	DeleteBucket func(string) error
	NextBucket   func() (string, error)
}

// NewFakeClient creates fake GCS client to be used in tests
func NewFakeClient(data []string) Client {
	index := 0
	remainingBuckets := append(data[:0:0], data...)
	return Client{
		Buckets: func() []string {
			return remainingBuckets
		},
		NextBucket: func() (string, error) {
			if len(data) < index+1 {
				return "", iterator.Done
			}
			bucketName := data[index]
			index++
			return bucketName, nil
		},
		DeleteBucket: func(bucketNameToDelete string) error {
			for i, bucketName := range remainingBuckets {
				if bucketName != bucketNameToDelete {
					continue
				}
				remainingBuckets = remainingBuckets[:i+copy(remainingBuckets[i:], remainingBuckets[i+1:])]
				return nil
			}
			return nil
		},
	}
}
