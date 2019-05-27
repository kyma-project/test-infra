package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func cleanFlags() {
	argProjectName = ""
	argBucketLifespanDuration = bucketLifespanDurationDefault
	argExcludedBucketNames = ""
	argDryRun = false
	argBucketNameRegexp = ""
	argBucketObjectWorkerNumber = bucketObjectWorkerNumberDefault
	argLogLevel = "info"
}

func TestConfigRead(t *testing.T) {

	const projectArgTag = "-project"
	const excludedBucketsArgTag = "-excludedBuckets"
	const durationArgTag = "-duration"
	const dryRunArgTag = "-dryRun"
	const bucketNameRegexpArgTag = "-bucketNameRegexp"
	const bucketObjectWorkerNumberTag = "-bucketObjectWorkerNumber"
	const logLevelTag = "-logLevel"

	tests := []struct {
		args                             []string
		expectedErr                      error
		expectedProjectName              string
		expectedDuration                 time.Duration
		expectedExcludedNames            []string
		expectedDryRun                   bool
		expectedBucketNameRegex          *regexp.Regexp
		expectedBucketObjectWorkerNumber int
		expectedLogLevel                 logrus.Level
	}{
		{
			args: []string{
				"cmd",
				projectArgTag, "test",
				bucketNameRegexpArgTag, "123",
				bucketObjectWorkerNumberTag, "10",
				logLevelTag, "debug",
			},
			expectedProjectName:              "test",
			expectedDuration:                 2 * time.Hour,
			expectedDryRun:                   false,
			expectedBucketNameRegex:          regexp.MustCompile("123"),
			expectedBucketObjectWorkerNumber: 10,
			expectedLogLevel:                 logrus.DebugLevel,
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test2",
				durationArgTag, "1m",
				dryRunArgTag,
				bucketNameRegexpArgTag, "123",
			},
			expectedProjectName:              "test2",
			expectedDuration:                 time.Minute,
			expectedDryRun:                   true,
			expectedBucketNameRegex:          regexp.MustCompile("123"),
			expectedBucketObjectWorkerNumber: bucketObjectWorkerNumberDefault,
			expectedLogLevel:                 logrus.InfoLevel,
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test3",
				durationArgTag, "1m",
				excludedBucketsArgTag, "test4,test5,test6",
				bucketNameRegexpArgTag, "123",
			},
			expectedProjectName:              "test3",
			expectedDuration:                 time.Minute,
			expectedExcludedNames:            []string{"test4", "test5", "test6"},
			expectedDryRun:                   false,
			expectedBucketNameRegex:          regexp.MustCompile("123"),
			expectedBucketObjectWorkerNumber: bucketObjectWorkerNumberDefault,
			expectedLogLevel:                 logrus.InfoLevel,
		},
		{
			args: []string{
				"cmd",
				bucketNameRegexpArgTag, "123",
			},
			expectedErr: ErrInvalidProjectName,
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "tes4",
				durationArgTag, "?",
				bucketNameRegexpArgTag, "123",
			},
			expectedErr: ErrInvalidDuration,
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test4",
			},
			expectedErr: ErrEmptyBucketNameRegexp,
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test5",
				bucketNameRegexpArgTag, "[aa)@",
			},
			expectedErr: ErrInvalidBucketNameRegexp,
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test5",
				bucketNameRegexpArgTag, "123",
				logLevelTag, "1213",
			},
			expectedErr: ErrInvalidLogLevel,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf(`test %d: args:%s`, i, test.args[1:])
		t.Run(testName, func(t *testing.T) {
			cleanFlags()
			os.Args = test.args
			config, err := readCfg()
			assert := assert.New(t)
			assert.Equal(test.expectedErr, err)
			assert.Equal(test.expectedProjectName, config.ProjectName)
			assert.Equal(test.expectedDuration, config.BucketLifespanDuration)
			assert.Equal(test.expectedExcludedNames, config.ExcludedBucketNames)
			assert.Equal(test.expectedDryRun, config.IsDryRun)
			assert.Equal(test.expectedBucketNameRegex, config.BucketNameRegexp)
			assert.Equal(test.expectedBucketObjectWorkerNumber, config.BucketObjectWorkersNumber)
			assert.Equal(test.expectedLogLevel, config.LogLevel)
		})
	}
}
