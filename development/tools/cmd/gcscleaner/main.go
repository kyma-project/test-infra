package main

import (
	"context"
	"flag"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/gcscleaner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Config structure aggregating application configuration arguments
type options struct {
	ProjectName               string
	BucketLifespanDuration    time.Duration
	ExcludedBucketNames       []string
	DryRun                    bool
	BucketNameRegexp          regexp.Regexp
	BucketObjectWorkersNumber int
	LogLevel                  logrus.Level
}

type bucketError struct {
	sync.Mutex
	err error
}

func main() {

	ctx := context.Background()

	options, err := readOptions()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "reading arguments"))
	}

	logrus.SetLevel(options.LogLevel)
	client, err := storage.NewClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	deleteBucket := func(bucketName string) error {
		bucketNamesChan := make(chan string)
		var bucketError bucketError
		rand.NewSource(time.Now().UnixNano())
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			defer close(bucketNamesChan)
			objects := client.Bucket(bucketName).Objects(ctx, nil)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					objectAttrs, err := objects.Next()
					if err == iterator.Done {
						return
					}
					if err != nil {
						logrus.Error(errors.Wrap(err, "iterating bucket objects"))
						bucketError.Lock()
						bucketError.err = err
						bucketError.Unlock()
						cancel()
						return
					}
					if objectAttrs == nil {
						logrus.Error(errors.Wrap(ErrInvalidBucketObjectAttributes, "iterating bucket objects"))
						bucketError.Lock()
						bucketError.err = ErrBucketDeletionCanceled
						bucketError.Unlock()
						cancel()
						return
					}
					bucketNamesChan <- objectAttrs.Name
				}
			}
		}()

		var waitGroup sync.WaitGroup
		waitGroup.Add(options.BucketObjectWorkersNumber)
		go func() {
			for i := 0; i < options.BucketObjectWorkersNumber; i++ {
				go func() {
					defer waitGroup.Done()
					for {
						select {
						case <-ctx.Done():
							return
						case objectName, ok := <-bucketNamesChan:
							if !ok {
								return
							}
							logrus.Debug("deleting: ", objectName)
							if options.DryRun {
								continue
							}
							if err := client.Bucket(bucketName).Object(objectName).Delete(ctx); err != nil {
								logrus.Error(errors.Wrap(err, "deleting bucket object"))
								bucketError.Lock()
								bucketError.err = ErrBucketDeletionCanceled
								bucketError.Unlock()
								return
							}
						default:
						}
					}
				}()
			}
		}()
		waitGroup.Wait()
		err := bucketError.err
		if err != nil {
			return err
		}
		if options.DryRun {
			return nil
		}
		return client.Bucket(bucketName).Delete(ctx)
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
	argProjectName                  string
	argExcludedBucketNames          string
	argBucketLifespanDuration       string
	argBucketNameRegexp             string
	argDryRun                       bool
	argBucketObjectWorkerNumber     int
	argLogLevel                     string
	bucketLifespanDurationDefault   = "2h"
	bucketObjectWorkerNumberDefault = 1
	logLevelDefault                 = "info"
	// ErrInvalidProjectName returned if project name argument is invalid
	ErrInvalidProjectName = errors.New("invalid project name argument")
	// ErrInvalidDuration returned if duration argument is invalid
	ErrInvalidDuration = errors.New("invalid duration argument")
	// ErrEmptyBucketNameRegexp returned if argBucketNameRegexp is empty
	ErrEmptyBucketNameRegexp = errors.New("empty bucketNameRegexp argument")
	// ErrInvalidBucketNameRegexp returned if argBucketNameRegexp is invalid
	ErrInvalidBucketNameRegexp = errors.New("invalid bucketNameRegexp argument")
	// ErrInvalidBucketObjectAttributes returned when bucket object has invalid attributes
	ErrInvalidBucketObjectAttributes = errors.New("bucket object has invalid attributes")
	// ErrInvalidLogLevel returned when argLogLevel is invalid
	ErrInvalidLogLevel = errors.New("bucket object has invalid attributes")
	// ErrBucketDeletionCanceled returned when deletion of bucket was canceled
	ErrBucketDeletionCanceled = errors.New("bucket deletion canceled")
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

	flag.IntVar(
		&argBucketObjectWorkerNumber,
		"bucketObjectWorkerNumber",
		bucketObjectWorkerNumberDefault,
		"the number of workers that will be used to delete bucket object")

	flag.StringVar(
		&argLogLevel,
		"logLevel",
		logLevelDefault,
		"logging level [panic|fatal|error|warn|warning|info|debug|trace]")
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

	logLevel, err := logrus.ParseLevel(argLogLevel)
	if err != nil {
		return options{}, ErrInvalidLogLevel
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
	options.BucketObjectWorkersNumber = argBucketObjectWorkerNumber
	options.LogLevel = logLevel
	return options, nil
}
