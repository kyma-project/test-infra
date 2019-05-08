package gcscleaner

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner/fake"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestExtractTimestamp(t *testing.T) {
	tests := []struct {
		name, bucketName string
		expected         func() *string
	}{
		{
			name:       "match: a1b2b3",
			bucketName: "matching-bucket-name-1b6dibbg2ogqo",
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
			bucketName: "not.matching.the.bucket.name-1b6dibbg2ogq@",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "no match #4",
			bucketName: "not.matching.the.bucket.name-_a1s2d34d12",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "match: 1111",
			bucketName: "matching.bucket.name-1111",
			expected: func() *string {
				result := "1111"
				return &result
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := extractTimestampSuffix(test.bucketName)
			assert.New(t).Equal(test.expected(), actual)
		})
	}
}

func TestClean(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	bucketNames, bucketsToDelete, protectedBuckets := getTestData()
	client := fake.NewFakeClient(bucketNames)
	err := Clean(context.Background(), Config{
		BucketLifespanDuration: time.Second,
		ExcludedBucketNames:    append([]string{"atx-prow2"}, protectedBuckets...),
	}, client)
	if err != nil {
		t.Error(err)
	}
	actualBucketNames, err := fake.GetBucketNames(client.Buckets(context.Background(), "test-project"))
	if err != nil {
		t.Error(err)
	}
	assert := assert.New(t)
	assert.Equal(len(bucketNames)-len(bucketsToDelete), len(actualBucketNames))
	for _, bucketName := range actualBucketNames {
		for _, bucketToDelete := range bucketsToDelete {
			assert.NotEqual(bucketToDelete, bucketName)
		}
	}
}

func getTestData() (bucketNames []string, namesOfBucketsToBeDeleted []string, protectedBucketNames []string) {

	duration := strconv.FormatInt(time.Now().Add(-3 * time.Hour).UnixNano(), 32)

	protectedBucketNames = []string{
		fmt.Sprintf(`protected-bucket-%s`, duration),
	}

	namesOfBucketsToBeDeleted = []string{
		fmt.Sprintf(`bucket-to-delete-%s`, duration),
		fmt.Sprintf(`bucket-to-delete2-%s`, duration),
		fmt.Sprintf(`bucket-to-delete3-%s`, duration),
	}

	bucketWithFutureName := fmt.Sprintf(
		`future-bucket-%s`,
		strconv.FormatInt(time.Now().Add(time.Hour).UnixNano(), 32))

	for _, slice := range [][]string{
		{
			"atx-prow2",
			"atx-prow3",
			"atx-",
			bucketWithFutureName,
		},
		protectedBucketNames,
		namesOfBucketsToBeDeleted,
	} {
		bucketNames = append(bucketNames, slice...)
	}
	return
}
