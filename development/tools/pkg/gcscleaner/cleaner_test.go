package gcscleaner_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"google.golang.org/api/iterator"

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
			bucketName: "matching-Bucket-Name-1b6dibbg2ogqo",
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
			bucketName: "not.matching.the.Bucket.Name-1b6dibbg2ogq@",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "no match #4",
			bucketName: "not.matching.the.Bucket.Name-_a1s2d34d12",
			expected: func() *string {
				return nil
			},
		},
		{
			name:       "match: 1111",
			bucketName: "matching.Bucket.Name-1111",
			expected: func() *string {
				result := "1111"
				return &result
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := gcscleaner.ExtractTimestampSuffix(test.bucketName, *regexp.MustCompile(`^.+-([a-z0-9]+$)`))
			assert.New(t).Equal(test.expected(), actual)
		})
	}
}

func TestClean(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	bucketNames, bucketsToDelete, protectedBuckets := getTestData()
	actualBucketNames := append([]string(nil), bucketNames...)
	err := gcscleaner.Clean(
		testNextBucketNameFunc(bucketNames),
		testDeleteBucket(&actualBucketNames),
		gcscleaner.NewConfig(
			*regexp.MustCompile(`^.+-([a-z0-9]+$)`),
			append([]string{"atx-prow2"}, protectedBuckets...),
			time.Second))

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

	duration := strconv.FormatInt(time.Now().Add(-3*time.Hour).UnixNano(), 32)

	protectedBucketNames = []string{
		fmt.Sprintf(`protected-Bucket-%s`, duration),
	}

	namesOfBucketsToBeDeleted = []string{
		fmt.Sprintf(`Bucket-to-delete-%s`, duration),
		fmt.Sprintf(`Bucket-to-delete2-%s`, duration),
		fmt.Sprintf(`Bucket-to-delete3-%s`, duration),
	}

	bucketWithFutureName := fmt.Sprintf(
		`future-Bucket-%s`,
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

func testDeleteBucket(remainingBuckets *[]string) gcscleaner.DeleteBucketFunc {
	return func(bucketNameToDelete string) error {
		for i, bucketName := range *remainingBuckets {
			if bucketName != bucketNameToDelete {
				continue
			}
			*remainingBuckets = append((*remainingBuckets)[:i], (*remainingBuckets)[i+1:]...)
			return nil
		}
		return nil
	}
}

func testNextBucketNameFunc(bucketNames []string) func() (string, error) {
	index := 0
	return func() (string, error) {
		if len(bucketNames) < index+1 {
			return "", iterator.Done
		}
		bucketName := bucketNames[index]
		index++
		return bucketName, nil
	}
}
