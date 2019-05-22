package main

import (
	"context"
	"flag"
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
}

func main() {

	logrus.SetLevel(logrus.DebugLevel)
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
			ctx, cancel := context.WithCancel(context.Background())
			bucketNamesChan := make(chan string)
			errorChan := make(chan error)
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
							logrus.Error(err)
							errorChan <- err
							cancel()
						}
						if objectAttrs == nil {
							logrus.Error(ErrInvalidBucketObjectAttributes)
							errorChan <- err
							cancel()
						}
						bucketNamesChan <- objectAttrs.Name
					}
				}
			}()

			var waitGroup sync.WaitGroup
			waitGroup.Add(options.BucketObjectWorkersNumber)
			go func() {
				defer close(errorChan)
				for i := 0; i < options.BucketObjectWorkersNumber; i++ {
					go func() {
						defer waitGroup.Done()
						for {
							select {
							case <-ctx.Done():
								return
							default:
							}
							select {
							case objectName, ok := <-bucketNamesChan:
								if !ok {
									return
								}
								logrus.Info("deleting ", bucketName, "::", objectName)
								if err := client.Bucket(bucketName).Object(objectName).Delete(ctx); err != nil {
									logrus.Error(err)
									errorChan <- err
									cancel()
								}
							default:
							}
						}
					}()
				}
			}()
			waitGroup.Wait()
			var allErrors []string
			for err := range errorChan {
				allErrors = append(allErrors, err.Error())
			}
			if len(allErrors) > 0 {
				return errors.Wrap(ErrDeletingBucketObjects, strings.Join(allErrors, "\n"))
			}
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
	argProjectName                  string
	argExcludedBucketNames          string
	argBucketLifespanDuration       string
	argBucketNameRegexp             string
	argDryRun                       bool
	argBucketObjectWorkerNumber     int
	bucketLifespanDurationDefault   = "2h"
	bucketObjectWorkerNumberDefault = 1
	dryRunDeleteBucket              = func(_ string) error {
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
	// ErrDeletingBucketObjects returned when bucket object deletion was unsuccessful
	ErrDeletingBucketObjects = errors.New("error while deleting bucket object")
	// ErrInvalidBucketObjectAttributes returned when bucket object has invalid attributes
	ErrInvalidBucketObjectAttributes = errors.New("bucket object has invalid attributes")
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
	options.BucketObjectWorkersNumber = argBucketObjectWorkerNumber
	return options, nil
}
