package gcscleaner

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

type DeleteBucket func(string) error

type NextBucket func() (string, error)

// Clean cleans up buckets created by Asset Store
func Clean(
	nextBucket NextBucket,
	deleteBucket DeleteBucket,
	excludedBucketNames []string,
	bucketLifespanDuration time.Duration) error {

	var result error
	for {
		bucketName, err := nextBucket()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if !shouldDeleteBucket(
			bucketName,
			excludedBucketNames,
			time.Now().UnixNano(),
			bucketLifespanDuration) {

			continue
		}
		if err := deleteBucket(bucketName); err != nil {
			logrus.Error(errors.Wrap(
				err,
				fmt.Sprintf(`"deleting bucket %s"`, bucketName)),
			)
			result = ErrWhileDelBuckets
		}
		logrus.Info(fmt.Sprintf(`bucket: '%s' deleted`, bucketName))
	}
	return result
}

func shouldDeleteBucket(
	bucketName string,
	excludedBucketNames []string,
	now int64,
	bucketLifespanDuration time.Duration) bool {

	for _, excludedBucketName := range excludedBucketNames {
		if excludedBucketName != bucketName {
			continue
		}
		return false
	}

	timestampSuffix := extractTimestampSuffix(bucketName)
	if timestampSuffix == nil {
		return false
	}
	timestamp, err := strconv.ParseInt(*timestampSuffix, 32, 0)
	if err != nil {
		logrus.Debug(fmt.Sprintf(
			"skipping bucket '%s', unable to parse timestamp",
			bucketName),
		)
		return false
	}
	if now-timestamp < int64(bucketLifespanDuration) {
		logrus.Debug(fmt.Sprintf(
			"bucket: '%s' is %s old and will not be deleted, the duration: '%s' was not exceeded",
			bucketName,
			time.Duration(now-timestamp),
			bucketLifespanDuration),
		)
		return false
	}
	logrus.Debug(fmt.Sprintf(`bucket: '%s' will be deleted`, bucketName))
	return true
}

func extractTimestampSuffix(name string) *string {

	submatch := regTimestampSuffix.FindSubmatch([]byte(name))
	if len(submatch) < 2 {
		return nil
	}
	result := string(submatch[1])
	return &result
}

// Config structure aggregating application configuration arguments
type Config struct {
	ProjectName            string
	BucketLifespanDuration time.Duration
	ExcludedBucketNames    []string
	DryRun                 bool
}

var regTimestampSuffix = regexp.MustCompile(`^.+-([a-z0-9]+$)`)

// ErrWhileDelBuckets returned when error occurred while deleting one or more buckets
var ErrWhileDelBuckets = errors.New("error while deleting bucket")
