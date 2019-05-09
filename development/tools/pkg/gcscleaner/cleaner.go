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

// DeleteBucketFunc deletes bucket with a given name
type DeleteBucketFunc func(string) error

// BucketNameGeneratorFunc returns next bucket name on each call
type BucketNameGeneratorFunc func() (string, error)

// Config cleaner configuration
type Config struct {
	ExcludedBucketNames    []string
	BucketLifespanDuration time.Duration
	RegTimestampSuffix     regexp.Regexp
}

// Clean cleans up buckets created by Asset Store
func Clean(
	generateNextBucketName BucketNameGeneratorFunc,
	deleteBucket DeleteBucketFunc,
	cfg Config) error {

	var result error
	for {
		bucketName, err := generateNextBucketName()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if !shouldDeleteBucket(bucketName, time.Now().UnixNano(), cfg) {
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

func shouldDeleteBucket(bucketName string, now int64, cfg Config) bool {

	for _, excludedBucketName := range cfg.ExcludedBucketNames {
		if excludedBucketName != bucketName {
			continue
		}
		return false
	}

	timestampSuffix := extractTimestampSuffix(bucketName, cfg.RegTimestampSuffix)
	if timestampSuffix == nil {
		logrus.Debug(fmt.Sprintf(
			"skipping bucket '%s', no timestamp",
			bucketName),
		)
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
	if now-timestamp < int64(cfg.BucketLifespanDuration) {
		logrus.Debug(fmt.Sprintf(
			"bucket: '%s' is %s old and will not be deleted, the duration: '%s' was not exceeded",
			bucketName,
			time.Duration(now-timestamp),
			cfg.BucketLifespanDuration),
		)
		return false
	}
	logrus.Debug(fmt.Sprintf(`bucket: '%s' will be deleted`, bucketName))
	return true
}

func extractTimestampSuffix(name string, regTimestampSuffix regexp.Regexp) *string {

	submatch := regTimestampSuffix.FindSubmatch([]byte(name))
	if len(submatch) < 2 {
		return nil
	}
	result := string(submatch[1])
	return &result
}

// NewConfig creates new cleaner configuration
func NewConfig(
	bucketNameRegexp regexp.Regexp,
	excludedBucketNames []string,
	bucketLifespanDuration time.Duration) Config {

	return Config{
		RegTimestampSuffix:     bucketNameRegexp,
		ExcludedBucketNames:    excludedBucketNames,
		BucketLifespanDuration: bucketLifespanDuration,
	}
}

// ErrWhileDelBuckets returned when error occurred while deleting one or more buckets
var ErrWhileDelBuckets = errors.New("error while deleting bucket")
