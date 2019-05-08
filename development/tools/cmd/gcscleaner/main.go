package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	cfg, err := readConfig()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "reading configuration"))
	}

	deleteBucket := func() gcscleaner.DeleteBucket {
		if cfg.DryRun {
			return dryRunDeleteBucket
		}
		return func(bucketName string) error {
			return client.Bucket(bucketName).Delete(ctx)
		}
	}()

	nextBucket := func() gcscleaner.NextBucket {
		bucketIterator := client.Buckets(ctx, cfg.ProjectName)
		return func() (string, error) {
			attrs, err := bucketIterator.Next()
			if err != nil {
				return "", err
			}
			return attrs.Name, nil
		}
	}()

	err = gcscleaner.Clean(
		nextBucket,
		deleteBucket,
		cfg.ExcludedBucketNames,
		cfg.BucketLifespanDuration)

	if err != nil {
		logrus.Fatal(errors.Wrap(err, "cleaning buckets"))
	}
}

var (
	argProjectName                string
	argExcludedBucketNames        string
	argBucketLifespanDuration     string
	bucketLifespanDurationDefault = "2h"
	argDryRun                     bool
	dryRunDeleteBucket            = func(_ string) error {
		return nil
	}

	// ErrInvalidProjectName returned if project name argument is invalid
	ErrInvalidProjectName = errors.New("invalid project name argument")
	// ErrInvalidDuration returned if duration argument is invalid
	ErrInvalidDuration = errors.New("invalid duration argument")
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
}

func readConfig() (gcscleaner.Config, error) {
	flag.Parse()

	if argProjectName == "" {
		return gcscleaner.Config{}, ErrInvalidProjectName
	}

	duration, err := time.ParseDuration(argBucketLifespanDuration)
	if err != nil {
		return gcscleaner.Config{}, ErrInvalidDuration
	}

	cfg := gcscleaner.Config{}
	cfg.ProjectName = argProjectName
	cfg.BucketLifespanDuration = duration
	if argExcludedBucketNames != "" {
		cfg.ExcludedBucketNames = strings.Split(
			argExcludedBucketNames,
			",")
	}
	cfg.DryRun = argDryRun

	return cfg, nil
}
