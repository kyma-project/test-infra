package main

import (
	"context"
	"flag"
	_ "math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	argProjectName                string
	argExcludedBucketNames        string
	argBucketLifespanDuration     string
	argBucketNameRegexp           string
	argDryRun                     bool
	argWorkersNumber              int
	argLogLevel                   string
	bucketLifespanDurationDefault = "2h"
	workerNumberDefault           = 1
	logLevelDefault               = "info"
	// ErrInvalidProjectName returned if project name argument is invalid
	ErrInvalidProjectName = errors.New("invalid project name argument")
	// ErrInvalidDuration returned if duration argument is invalid
	ErrInvalidDuration = errors.New("invalid duration argument")
	// ErrEmptyBucketNameRegexp returned if argBucketNameRegexp is empty
	ErrEmptyBucketNameRegexp = errors.New("empty bucketNameRegexp argument")
	// ErrInvalidBucketNameRegexp returned if argBucketNameRegexp is invalid
	ErrInvalidBucketNameRegexp = errors.New("invalid bucketNameRegexp argument")
	// ErrInvalidLogLevel returned when argLogLevel is invalid
	ErrInvalidLogLevel = errors.New("bucket object has invalid attributes")
)

func main() {
	rootCtx := context.Background()
	cfg, err := readCfg()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "while reading arguments"))
	}
	logrus.SetLevel(cfg.LogLevel)
	client, err := storage.NewClient(rootCtx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer client.Close()
	cleaner := gcscleaner.NewCleaner(client, cfg)
	err = cleaner.DeleteOldBuckets(rootCtx)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "while deleting old buckets"))
	}
}

func readCfg() (gcscleaner.Config, error) {
	flag.Parse()

	if argBucketNameRegexp == "" {
		return gcscleaner.Config{}, ErrEmptyBucketNameRegexp
	}

	bucketNameRegexp, err := regexp.Compile(argBucketNameRegexp)
	if err != nil {
		return gcscleaner.Config{}, ErrInvalidBucketNameRegexp
	}

	if argProjectName == "" {
		return gcscleaner.Config{}, ErrInvalidProjectName
	}

	duration, err := time.ParseDuration(argBucketLifespanDuration)
	if err != nil {
		return gcscleaner.Config{}, ErrInvalidDuration
	}

	logLevel, err := logrus.ParseLevel(argLogLevel)
	if err != nil {
		return gcscleaner.Config{}, ErrInvalidLogLevel
	}

	cfg := gcscleaner.Config{}
	cfg.ProjectName = argProjectName
	cfg.BucketLifespanDuration = duration
	if argExcludedBucketNames != "" {
		cfg.ExcludedBucketNames = strings.Split(
			argExcludedBucketNames,
			",")
	}
	cfg.IsDryRun = argDryRun
	cfg.BucketNameRegexp = bucketNameRegexp
	cfg.BucketObjectWorkersNumber = argWorkersNumber
	cfg.LogLevel = logLevel
	return cfg, nil
}

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

	flag.IntVar(
		&argWorkersNumber,
		"workerNumber",
		workerNumberDefault,
		"the number of workers that will be used to delete bucket object")

	flag.StringVar(
		&argLogLevel,
		"logLevel",
		logLevelDefault,
		"logging level [panic|fatal|error|warn|warning|info|debug|trace]")
}
