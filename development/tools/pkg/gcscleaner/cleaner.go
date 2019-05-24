package gcscleaner

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
)

// DeleteBucketFunc deletes Bucket with a given Name
type DeleteBucketFunc func(string) error

// BucketNameGeneratorFunc returns next Bucket Name on each call
type BucketNameGeneratorFunc func() (string, error)

// NewConfig creates new cleaner configuration
func NewConfig(
	bucketNameRegexp regexp.Regexp,
	excludedBucketNames []string,
	bucketLifespanDuration time.Duration) Config {

	return Config{
		BucketNameRegexp:       bucketNameRegexp,
		ExcludedBucketNames:    excludedBucketNames,
		BucketLifespanDuration: bucketLifespanDuration,
	}
}

// ErrWhileDelBuckets returned when error occurred while deleting one or more buckets
var ErrWhileDelBuckets = errors.New("error while deleting Bucket")
