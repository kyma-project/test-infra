package gcscleaner

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
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
	protectedBucketName := fmt.Sprintf(`protected-bucket-to-delete-%s`, strconv.FormatInt(time.Now().Add(-3 * time.Hour).UnixNano(), 32))
	bucket2Delete := fmt.Sprintf(`bucket-to-delete-%s`, strconv.FormatInt(time.Now().Add(-3 * time.Hour).UnixNano(), 32))
	defaultBucketNames := []string{
		"atx-prow2",
		bucket2Delete,
		"atx-prow3",
		"atx-",
		protectedBucketName,
		fmt.Sprintf(`future-bucket-%s`, strconv.FormatInt(time.Now().Add(time.Hour).UnixNano(), 32)),
	}
	client := NewFakeClient(defaultBucketNames)
	err := Clean(context.Background(), Config{
		BucketLifespanDuration: time.Second,
		ExcludedBucketNames:    []string{"atx-prow2", protectedBucketName},
	}, client)
	if err != nil {
		t.Error(err)
	}
	actualBucketNames, err := GetBucketNames(client.Buckets(context.Background(), "test-project"))
	if err != nil {
		t.Error(err)
	}
	assert := assert.New(t)
	assert.Equal(len(defaultBucketNames)-1, len(actualBucketNames))
	for _, bucketName := range actualBucketNames {
		assert.NotEqual(bucket2Delete, bucketName)
	}
}
