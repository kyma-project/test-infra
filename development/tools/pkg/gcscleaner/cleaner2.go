package gcscleaner

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BucketObject struct {
	Name   string
	Bucket string
}

// Config cleaner configuration
type Config struct {
	ProjectName               string
	BucketLifespanDuration    time.Duration
	ExcludedBucketNames       []string
	IsDryRun                  bool
	BucketNameRegexp          regexp.Regexp
	BucketObjectWorkersNumber int
	LogLevel                  logrus.Level // TODO do not pass it into cleaner
}

//go:generate mockery -name=BucketAttrs -output=automock -outpkg=automock -case=underscore
type BucketAttrs interface {
	Name() string
}

type bucketAttrs struct {
	bucketAttrs *storage.BucketAttrs
}

func (r bucketAttrs) Name() string {
	return r.bucketAttrs.Name
}

//go:generate mockery -name=Query -output=automock -outpkg=automock -case=underscore
type Query interface {
	Delimiter() string
	Prefix() string
	Versions() bool
}

//go:generate mockery -name=ObjectAttrs -output=automock -outpkg=automock -case=underscore
type ObjectAttrs interface {
	Name() string
	Bucket() string
}

//go:generate mockery -name=ObjectIterator -output=automock -outpkg=automock -case=underscore
type ObjectIterator interface {
	Next() (ObjectAttrs, error)
}

//go:generate mockery -name=ObjectHandle -output=automock -outpkg=automock -case=underscore
type ObjectHandle interface {
	Delete(ctx context.Context) error
}

//go:generate mockery -name=BucketHandle -output=automock -outpkg=automock -case=underscore
type BucketHandle interface {
	Object(name string) ObjectHandle
	Objects(ctx context.Context, q Query) ObjectIterator
	Delete(ctx context.Context) (err error)
}

//go:generate mockery -name=BucketIterator -output=automock -outpkg=automock -case=underscore
type BucketIterator interface {
	Next() (BucketAttrs, error)
}

type bucketIterator struct {
	bucketIterator *storage.BucketIterator
}

func (r bucketIterator) Next() (BucketAttrs, error) {
	storageBucketAttrs, err := r.bucketIterator.Next()
	if err != nil {
		return nil, err
	}
	bucketAttrs := bucketAttrs{
		bucketAttrs: storageBucketAttrs,
	}
	return bucketAttrs, nil
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Bucket(bucketName string) BucketHandle
	Buckets(ctx context.Context, projectID string) BucketIterator
	Close() error
}

type CancelableContext struct {
	context.Context
	Cancel func()
}

func NewCancelableContext(ctx context.Context) CancelableContext {
	cancelableCtx, cancel := context.WithCancel(ctx)
	return CancelableContext{
		Context: cancelableCtx,
		Cancel:  cancel,
	}
}

type Cleaner2 struct {
	ctx    context.Context
	client Client
	cfg    Config
}

// NewCleaner2 creates cleaner
func NewCleaner2(client Client, cfg Config) Cleaner2 {
	return Cleaner2{
		client: client,
		cfg:    cfg,
	}
}

func (r Cleaner2) shouldDeleteBucket(bucketName string, now int64) bool {
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

func (r Cleaner2) extractTimestampSuffix(name string) *string {
	submatch := r.cfg.BucketNameRegexp.FindSubmatch([]byte(name))
	if len(submatch) < 2 {
		return nil
	}
	result := string(submatch[1])
	return &result
}

func (r Cleaner2) deleteBucketObject(
	ctx context.Context,
	bucketName string,
	objectName string) error {
	if r.cfg.IsDryRun {
		msg := fmt.Sprintf(`dry-run|delete object: %s`, objectName) // FIXME change it
		logrus.Debug(msg)
		return nil
	}
	err := r.client.Bucket(bucketName).Object(objectName).Delete(ctx)
	logrus.Debug("object deleted: ", objectName)
	return err
}

func (r Cleaner2) iterateBucketObjectNames(
	ctx context.Context,
	bucketName string,
	bucketObjectChan chan BucketObject,
	errChan chan error) {
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
				errChan <- err
				return
			}
			bucketObjectChan <- BucketObject{
				Bucket: attrs.Bucket(),
				Name:   attrs.Name(),
			}
		}
	}
}

func (r Cleaner2) deleteBucketObjects(
	ctx CancelableContext,
	bucketObjectChan chan BucketObject,
	errChan chan error) {
	for {
		select {
		case bo, ok := <-bucketObjectChan:
			if !ok {
				return
			}
			if err := r.deleteBucketObject(ctx, bo.Bucket, bo.Name); err != nil {
				errChan <- err
				ctx.Cancel()
				return
			}
		default:
		}
	}
}

// deleteBucketObjects deletes old buckets within GCP project
func (r Cleaner2) DeleteOldBuckets(rootCtx context.Context) error {
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

func (r Cleaner2) deleteBucket(ctx CancelableContext, bucketName string) error {
	errChan := make(chan error)
	go r.deleteAllObjects(ctx, bucketName, errChan)
	err := r.collectErrors(errChan)
	if err != nil {
		return err
	}
	if r.cfg.IsDryRun {
		logrus.Info(`deleted bucket: `, bucketName)
		return nil
	}
	err = r.client.Bucket(bucketName).Delete(ctx)
	if err != nil {
		message := fmt.Sprintf(`deleting bucket %s`, bucketName)
		return errors.Wrap(err, message)
	}
	return nil
}

func (r Cleaner2) collectErrors(errChan chan error) error {
	var errorMessages []string
	for {
		select {
		case err, ok := <-errChan:
			if !ok {
				return r.parseErrors(errorMessages)
			}
			errorMessages = append(errorMessages, err.Error())
		default:
		}
	}
}

func (r Cleaner2) parseErrors(errorMessages []string) error {
	if len(errorMessages) == 0 {
		return nil
	}
	errorMessage := strings.Join(errorMessages, "\n")
	return fmt.Errorf(errorMessage)
}
