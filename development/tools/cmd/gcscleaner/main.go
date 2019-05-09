package main

import (
	"context"
	"flag"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Config structure aggregating application configuration arguments
type options struct {
	ProjectName            string
	BucketLifespanDuration time.Duration
	ExcludedBucketNames    []string
	DryRun                 bool
	BucketNameRegexp       regexp.Regexp
}

func main() {
	ctx := context.Background()

	options, err := readOptions()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "reading arguments"))
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	var deleteBucket gcscleaner.DeleteBucketFunc
	if options.DryRun {
		deleteBucket = dryRunDeleteBucket
	} else {
		deleteBucket = func(bucketName string) error {
			return client.Bucket(bucketName).Delete(ctx)
		}
	}

	bucketIterator := client.Buckets(ctx, options.ProjectName)
	nextBucketNameGenerator := func() (string, error) {
		attrs, err := bucketIterator.Next()
		if err != nil {
			return "", err
		}
		return attrs.Name, nil
	}

	err = gcscleaner.Clean(
		nextBucketNameGenerator,
		deleteBucket,
		gcscleaner.NewConfig(
			options.BucketNameRegexp,
			options.ExcludedBucketNames,
			options.BucketLifespanDuration))

	if err != nil {
		logrus.Fatal(errors.Wrap(err, "cleaning buckets"))
	}
}

var (
	argProjectName                string
	argExcludedBucketNames        string
	argBucketLifespanDuration     string
	argBucketNameRegexp           string
	argDryRun                     bool
	bucketLifespanDurationDefault = "2h"
	dryRunDeleteBucket            = func(_ string) error {
		return nil
	}
	// ErrInvalidProjectName returned if project name argument is invalid
	ErrInvalidProjectName = errors.New("invalid project name argument")
	// ErrInvalidDuration returned if duration argument is invalid
	ErrInvalidDuration = errors.New("invalid duration argument")
	// ErrEmptyBucketNameRegexp returned if argBucketNameRegexp is empty
	ErrEmptyBucketNameRegexp = errors.New("empty bucketNameRegexp argument")
	// ErrInvalidBucketNameRegexp returned if argBucketNameRegexp is invalid
	ErrInvalidBucketNameRegexp = errors.New("invalid bucketNameRegexp argument")
)

func init() {
	flag.StringVar(
		&argProjectName,
		"project",
		"",
		"google cloud project name")

	flag.StringVar(
		&argBucketLifespanDuration,
		"duration",
		bucketLifespanDurationDefault,
		"buckt lifespan duration",
	)

	flag.StringVar(
		&argExcludedBucketNames,
		"excludedBuckets",
		"",
		"bucket names that are protected from deletion")

	flag.BoolVar(
		&argDryRun,
		"dryRun",
		false,
		"dry Run enabled, nothing is deleted")

	flag.StringVar(
		&argBucketNameRegexp,
		"bucketNameRegexp",
		"",
		"bucket name regexp pattern used to mach when deleted buckets")
}

func readOptions() (options, error) {
	flag.Parse()

	if argBucketNameRegexp == "" {
		return options{}, ErrEmptyBucketNameRegexp
	}

	bucketNameRegexp, err := regexp.Compile(argBucketNameRegexp)
	if err != nil {
		return options{}, ErrInvalidBucketNameRegexp
	}

	if argProjectName == "" {
		return options{}, ErrInvalidProjectName
	}

	duration, err := time.ParseDuration(argBucketLifespanDuration)
	if err != nil {
		return options{}, ErrInvalidDuration
	}

	options := options{}
	options.ProjectName = argProjectName
	options.BucketLifespanDuration = duration
	if argExcludedBucketNames != "" {
		options.ExcludedBucketNames = strings.Split(
			argExcludedBucketNames,
			",")
	}
	options.DryRun = argDryRun
	options.BucketNameRegexp = *bucketNameRegexp

	return options, nil
}
