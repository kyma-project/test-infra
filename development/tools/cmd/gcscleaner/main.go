package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
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

	err = gcscleaner.Clean(ctx, cfg, stiface.AdaptClient(client))
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "cleaning buckets"))
	}
}

var (
	argProjectName                string
	argExcludedBucketNames        string
	argBucketLifespanDuration     string
	bucketLifespanDurationDefault = "2h"

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

	return cfg, nil
}
