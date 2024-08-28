package gcscleaner

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	storage2 "github.com/kyma-project/test-infra/pkg/tools/gcscleaner/storage"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

// Config cleaner configuration
type Config struct {
	ProjectName               string
	BucketLifespanDuration    time.Duration
	ExcludedBucketNames       []string
	IsDryRun                  bool
	BucketNameRegexp          *regexp.Regexp
	BucketObjectWorkersNumber int
	LogLevel                  logrus.Level
}

// CancelableContext contains both context and it's Cancel function
type CancelableContext struct {
	context.Context
	Cancel func()
}

// NewCancelableContext creates new CancelableContext
func NewCancelableContext(ctx context.Context) CancelableContext {
	cancelableCtx, cancel := context.WithCancel(ctx)
	return CancelableContext{
		Context: cancelableCtx,
		Cancel:  cancel,
	}
}

// Cleaner cleans GCP buckets
type Cleaner struct {
	client storage2.Client
	cfg    Config
}

// NewCleaner creates cleaner
func NewCleaner(client storage2.Client, cfg Config) Cleaner {
	return Cleaner{
		client: client,
		cfg:    cfg,
	}
}

func (r Cleaner) shouldDeleteBucket(bucketName string, now int64) bool {
	for _, excludedBucketName := range r.cfg.ExcludedBucketNames {
		if excludedBucketName != bucketName {
			continue
		}
		logrus.Info(fmt.Sprintf(
			"skipping bucket '%s', bucket is excluded",
			bucketName),
		)
		return false
	}
	timestampSuffix := r.extractTimestampSuffix(bucketName)
	if timestampSuffix == nil {
		logrus.Info(fmt.Sprintf(
			"skipping bucket '%s', no timestamp",
			bucketName),
		)
		return false
	}
	timestamp, err := strconv.ParseInt(*timestampSuffix, 32, 0)
	if err != nil {
		logrus.Info(fmt.Sprintf(
			"skipping bucket '%s', unable to parse timestamp",
			bucketName),
		)
		return false
	}
	duration := now - timestamp
	if duration < int64(r.cfg.BucketLifespanDuration) {
		logrus.Info(fmt.Sprintf(
			"bucket: '%s' is %s old and will not be deleted, the duration: '%s' was not exceeded",
			bucketName,
			time.Duration(duration),
			r.cfg.BucketLifespanDuration),
		)
		return false
	}
	logrus.Debug(fmt.Sprintf(`bucket: '%s' was created %s ago and will be deleted`, bucketName, time.Duration(duration)))
	return true
}

func (r Cleaner) extractTimestampSuffix(name string) *string {
	submatch := r.cfg.BucketNameRegexp.FindSubmatch([]byte(name))
	if len(submatch) < 2 {
		return nil
	}
	result := string(submatch[1])
	return &result
}

func (r Cleaner) deleteBucketObject(
	ctx context.Context,
	bucketName string,
	objectName string) error {
	msg := fmt.Sprintf("object deleted: %s", objectName)
	if r.cfg.IsDryRun {
		logrus.Debug("[dry-run] ", msg)
		return nil
	}
	err := r.client.Bucket(bucketName).Object(objectName).Delete(ctx)
	logrus.Debug(msg)
	return err
}

func (r Cleaner) iterateBucketObjectNames(ctx context.Context, bucketName string, bucketObjectChan chan storage2.BucketObject, errChan chan error) {
	defer close(bucketObjectChan)

	bucket := r.client.Bucket(bucketName)
	objectIterator := bucket.Objects(ctx, nil)
	for {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
			attrs, err := objectIterator.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				errChan <- errors.Wrap(err, "while iterating bucket object names")
				return
			}
			bucketObjectChan <- storage2.NewBucketObject(attrs.Bucket(), attrs.Name())
		}
	}
}

func (r Cleaner) deleteBucketObjects(ctx CancelableContext, bucketObjectChan chan storage2.BucketObject, errChan chan error) {
	for bo := range bucketObjectChan {
		if err := r.deleteBucketObject(ctx, bo.Bucket(), bo.Name()); err != nil {
			errChan <- errors.Wrap(err, "while deleting bucket object")
			ctx.Cancel()
			return
		}
	}
}

// DeleteOldBuckets deletes old buckets within GCP project
func (r Cleaner) DeleteOldBuckets(rootCtx context.Context) error {
	bucketIterator := r.client.Buckets(rootCtx, r.cfg.ProjectName)
	var errorMessages []string
	for {
		bucketAttrs, err := bucketIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if !r.shouldDeleteBucket(bucketAttrs.Name(), time.Now().UnixNano()) {
			continue
		}
		cancelableCtx := NewCancelableContext(rootCtx)
		err = r.deleteBucket(cancelableCtx, bucketAttrs.Name())
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}
	return r.parseErrors(errorMessages)
}

func (r Cleaner) deleteBucket(ctx CancelableContext, bucketName string) error {
	errChan := make(chan error)
	go r.deleteAllObjects(ctx, bucketName, errChan)
	err := r.collectErrors(errChan)
	if err != nil {
		return err
	}
	if r.cfg.IsDryRun {
		logrus.Info(`[dry-run] deleted bucket: `, bucketName)
		return nil
	}
	err = r.client.Bucket(bucketName).Delete(ctx)
	if err != nil {
		message := fmt.Sprintf(`deleting bucket %s`, bucketName)
		return errors.Wrap(err, message)
	}
	return nil
}

func (r Cleaner) collectErrors(errChan chan error) error {
	var errorMessages []string
	for err := range errChan {
		errorMessages = append(errorMessages, err.Error())
	}
	return r.parseErrors(errorMessages)
}

func (r Cleaner) parseErrors(errorMessages []string) error {
	if len(errorMessages) == 0 {
		return nil
	}
	errorMessage := strings.Join(errorMessages, "\n")
	return fmt.Errorf(errorMessage) //nolint:govet
}

func (r Cleaner) deleteAllObjects(ctx CancelableContext, bucketName string, errChan chan error) {
	defer close(errChan)

	bucketObjectChan := make(chan storage2.BucketObject)
	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		r.iterateBucketObjectNames(ctx, bucketName, bucketObjectChan, errChan)
	}()

	for i := 0; i < r.cfg.BucketObjectWorkersNumber; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			r.deleteBucketObjects(ctx, bucketObjectChan, errChan)
		}()
	}
	waitGroup.Wait()
}
