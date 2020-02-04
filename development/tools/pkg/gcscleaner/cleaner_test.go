package gcscleaner_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner/storage/automock"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
)

func TestExtractTimestamp(t *testing.T) {
	tests := []struct {
		name, bucketName string
		expected         func() *string
	}{
		{
			name:       "match: a1b2b3",
			bucketName: "matching-Bucket-name-1b6dibbg2ogqo",
			expected: func() *string {
				result := "1b6dibbg2ogqo"
				return &result
			},
		},
		{
			name:       "no match #1",
			bucketName: "-a1s2d34d11",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "no match #2",
			bucketName: "a1s2d34d11",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "no match #3",
			bucketName: "not.matching.the.Bucket.name-1b6dibbg2ogq@",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "no match #4",
			bucketName: "not.matching.the.Bucket.name-_a1s2d34d12",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "match: 1111",
			bucketName: "matching.Bucket.name-1111",
			expected: func() *string {
				result := "1111"
				return &result
			},
		},
	}
	cleaner := newTestCleaner(&automock.Client{}, nil)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := cleaner.ExtractTimestampSuffix(test.bucketName)
			assert.New(t).Equal(test.expected(), actual)
		})
	}
}

func TestCleaner_ShouldDeleteBucket(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	now := time.Now()
	excludedBucketName := fmt.Sprintf(`excluded-bucket-%s`, strconv.FormatInt(now.Add(-3*time.Hour).UnixNano(), 32))
	cleaner := newTestCleaner(&automock.Client{}, []string{
		excludedBucketName,
	})
	tests := []struct {
		name     string
		expected bool
	}{
		{
			name:     fmt.Sprintf(`old-bucket-to-delete-%s`, strconv.FormatInt(now.Add(-3*time.Hour).UnixNano(), 32)),
			expected: true,
		},
		{
			name:     fmt.Sprintf(`excluded-bucket-%s`, strconv.FormatInt(now.Add(-3*time.Hour).UnixNano(), 32)),
			expected: false,
		},
		{
			name:     fmt.Sprintf(`bucket-not-to-delete-%s`, strconv.FormatInt(now.UnixNano(), 32)),
			expected: false,
		},
		{
			name:     "make-sure-to-exclude-cause-this-one-wil-be-deleted",
			expected: true,
		},
		{
			name:     fmt.Sprintf(`bucket-will-not-be-deleted_%s`, strconv.FormatInt(now.Add(-3*time.Hour).UnixNano(), 32)),
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := cleaner.ShouldDeleteBucket(test.name, now.UnixNano())
			assert := gomega.NewWithT(t)
			assert.Expect(actual).To(gomega.Equal(test.expected))
		})
	}
}

func newTestCleaner(client storage.Client, excludedBucketNames []string) gcscleaner.Cleaner {
	return gcscleaner.NewCleaner(client, getConf(excludedBucketNames))
}

func TestCleaner_DeleteOldBuckets_ErrBucketIteration(t *testing.T) {

	logrus.SetLevel(logrus.DebugLevel)
	errTest := errors.New("error while iterating buckets")

	bucketIterator := automock.BucketIterator{}
	bucketIterator.
		On("Next").
		Return(nil, errTest).
		Once()

	client := automock.Client{}
	client.
		On("Buckets", mock.AnythingOfType("*context.emptyCtx"), "test-project").
		Return(&bucketIterator).
		Once()
	cleaner := newTestCleaner(&client, nil)
	err := cleaner.DeleteOldBuckets(context.Background())
	assert := gomega.NewWithT(t)
	assert.Expect(err).To(gomega.Equal(errTest))
}

func TestCleaner_DeleteOldBuckets_ErrDeleteObject(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	now := time.Now()
	bucketIterator, bucketName := getBucketTestData(now)
	client := getStorageClient(bucketIterator)
	objectAttrs := automock.ObjectAttrs{}
	testObjectName := "test-object"
	objectIterator := getTestObjectIterator(&objectAttrs, testObjectName, bucketName)
	objectHandle := automock.ObjectHandle{}
	errTestDeleteObject := errors.New("error while deleting object")
	objectHandle.
		On("Delete", mock.Anything).
		Return(errTestDeleteObject).
		Once()
	bucketHandle := getTestBucketHandle(objectIterator, testObjectName, &objectHandle)
	client.
		On("Bucket", bucketName).
		Return(bucketHandle)
	cleaner := newTestCleaner(client, nil)
	ctx := context.Background()
	err := cleaner.DeleteOldBuckets(ctx)
	assert := gomega.NewWithT(t)
	assert.Expect(err).NotTo(gomega.BeNil())
}

func getTestObjectIterator(attrs *automock.ObjectAttrs, testObjectName string, bucketName string) *automock.ObjectIterator {
	attrs.
		On("name").
		Return(testObjectName).
		Once().
		On("Bucket").
		Return(bucketName).
		Once()

	objectIterator := automock.ObjectIterator{}
	objectIterator.
		On("Next").
		Return(attrs, nil).
		Once().
		On("Next").
		Return(nil, iterator.Done).
		Once()
	return &objectIterator
}

func TestCleaner_DeleteOldBuckets(t *testing.T) {

	var delObjectCounter int32

	logrus.SetLevel(logrus.DebugLevel)
	now := time.Now()
	bucketIterator, bucketName := getBucketTestData(now)
	client := getStorageClient(bucketIterator)

	objectAttrs := automock.ObjectAttrs{}
	testObjectName := "test-object"
	objectIterator := getTestObjectIterator(&objectAttrs, testObjectName, bucketName)

	objectHandle := automock.ObjectHandle{}
	objectHandle.
		On("Delete", mock.Anything).
		Run(func(_ mock.Arguments) {
			atomic.AddInt32(&delObjectCounter, 1)
		}).
		Return(nil).
		Once()

	bucketHandle := getTestBucketHandle(objectIterator, testObjectName, &objectHandle)

	client.
		On("Bucket", bucketName).
		Return(bucketHandle)

	cleaner := newTestCleaner(client, nil)
	ctx := context.Background()
	err := cleaner.DeleteOldBuckets(ctx)
	assert := gomega.NewWithT(t)
	assert.Expect(err).To(gomega.BeNil())
	assert.Expect(delObjectCounter).To(gomega.Equal(int32(1)))
}

func getTestBucketHandle(objectIterator *automock.ObjectIterator, testObjectName string, objectHandle *automock.ObjectHandle) *automock.BucketHandle {
	bucketHandle := automock.BucketHandle{}
	bucketHandle.
		On("Objects", mock.AnythingOfType("gcscleaner.CancelableContext"), nil).
		Return(objectIterator).
		Once().
		On("Object", testObjectName).
		Return(objectHandle).
		Once().
		On("Delete", mock.AnythingOfType("gcscleaner.CancelableContext")).
		Return(nil).
		Once()
	return &bucketHandle
}

func getStorageClient(bucketIterator storage.BucketIterator) *automock.Client {
	client := automock.Client{}
	client.
		On("Buckets", mock.AnythingOfType("*context.emptyCtx"), "test-project").
		Return(bucketIterator).
		Once()
	return &client
}

func getConf(excludedBucketNames []string) gcscleaner.Config {
	return gcscleaner.Config{
		BucketLifespanDuration:    2 * time.Hour,
		ProjectName:               "test-project",
		BucketNameRegexp:          regexp.MustCompile("^.+-([a-z0-9]+$)"),
		BucketObjectWorkersNumber: 1,
		ExcludedBucketNames:       excludedBucketNames,
		IsDryRun:                  false,
	}
}

func getBucketTestData(t time.Time) (storage.BucketIterator, string) {
	bucketIterator := automock.BucketIterator{}

	// bucket to be deleted
	formatInt := strconv.FormatInt(t.Add(-3*time.Hour).UnixNano(), 32)
	bucketName := fmt.Sprintf(`test-bucket-to-be-deleted-%s`, formatInt)
	b1 := automock.BucketAttrs{}
	b1.On("name").Return(bucketName)

	bucketIterator.
		On("Next").
		Return(&b1, nil).
		Once().
		On("Next").
		Return(nil, iterator.Done).
		Once()

	return &bucketIterator, bucketName
}
