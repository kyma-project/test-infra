package gcscleaner

import (
	"context"
	"fmt"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"regexp"
	"strconv"
	"time"
)

func Clean(
	ctx context.Context,
	cfg Config,
	client stiface.Client) error {

	buckets := client.Buckets(ctx, cfg.ProjectName)
	var result error
	for {
		attrs, err := buckets.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if !shouldDeleteBucket(
			attrs.Name,
			cfg.ExcludedBucketNames,
			time.Now().UnixNano(),
			cfg.BucketLifespanDuration) {

			continue
		}
		if err := client.Bucket(attrs.Name).Delete(ctx); err != nil {
			logrus.Error(errors.Wrap(
				err,
				fmt.Sprintf(`"deleting bucket %s"`, attrs.Name)),
			)
			result = ErrWhileDelBuckets
		}
		logrus.Info(fmt.Sprintf(`bucket: '%s' deleted`, attrs.Name))
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

type Config struct {
	ProjectName            string
	BucketLifespanDuration time.Duration
	ExcludedBucketNames    []string
}

var regTimestampSuffix = regexp.MustCompile(`^.+-([a-z0-9]+$)`)

var ErrWhileDelBuckets = errors.New("error while deleting bucket")
