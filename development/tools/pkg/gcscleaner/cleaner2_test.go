package gcscleaner_test

import (
	"context"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner/automock"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
	"testing"
)

func TestMe(t *testing.T) {

	client := &automock.Client{}
	cleaner2 := gcscleaner.NewCleaner2(client)

	errChan := make(chan error)
	bucketObjectChan := make(chan gcscleaner.BucketObject)
	bucketHandle := &automock.BucketHandle{}
	objectIterator := &automock.ObjectIterator{}
	objectIterator.On("Next").Return(nil, iterator.Done).Once()

	cancelableContext := gcscleaner.NewCancelableContext(context.Background())
	bucketHandle.On("Objects", mock.AnythingOfType("gcscleaner.CancelableContext"), nil).Return(objectIterator).Once()

	client.On("Bucket", "testbucket").Return(bucketHandle).Once()
	cleaner2.BucketObjectNamesChan(cancelableContext, "testbucket", bucketObjectChan, errChan)

}
