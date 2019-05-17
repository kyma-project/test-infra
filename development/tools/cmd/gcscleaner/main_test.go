package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func cleanFlags() {
	argProjectName = ""
	argBucketLifespanDuration = bucketLifespanDurationDefault
	argExcludedBucketNames = ""
	argDryRun = false
	argBucketNameRegexp = ""
}

func TestConfigRead(t *testing.T) {

	const projectArgTag = "-project"
	const excludedBucketsArgTag = "-excludedBuckets"
	const durationArgTag = "-duration"
	const dryRunArgTag = "-dryRun"
	const bucketNameRegexpArgTag = "-bucketNameRegexp"

	tests := []struct {
		args                    []string
		expectedErr             error
		expectedProjectName     string
		expectedDuration        time.Duration
		expectedExcludedNames   []string
		expectedDryRun          bool
		expectedBucketNameRegex regexp.Regexp
	}{
		{
			args: []string{
				"cmd",
				projectArgTag, "test",
				bucketNameRegexpArgTag, "123",
			},
			expectedProjectName:     "test",
			expectedDuration:        2 * time.Hour,
			expectedDryRun:          false,
			expectedBucketNameRegex: *regexp.MustCompile("123"),
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test2",
				durationArgTag, "1m",
				dryRunArgTag,
				bucketNameRegexpArgTag, "123",
			},
			expectedProjectName:     "test2",
			expectedDuration:        time.Minute,
			expectedDryRun:          true,
			expectedBucketNameRegex: *regexp.MustCompile("123"),
		},
		{
			args: []string{
				"cmd",
				projectArgTag, "test3",
				durationArgTag, "1m",
				excludedBucketsArgTag, "test4,test5,test6",
				bucketNameRegexpArgTag, "123",
			},
			expectedProjectName:     "test3",
			expectedDuration:        time.Minute,
			expectedExcludedNames:   []string{"test4", "test5", "test6"},
			expectedDryRun:          false,
			expectedBucketNameRegex: *regexp.MustCompile("123"),
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
	}
	for i, test := range tests {
		testName := fmt.Sprintf(`test %d: args:%s`, i, test.args[1:])
		t.Run(testName, func(t *testing.T) {
			cleanFlags()
			os.Args = test.args
			options, err := readOptions()
			assert := assert.New(t)
			assert.Equal(test.expectedErr, err)
			assert.Equal(test.expectedProjectName, options.ProjectName)
			assert.Equal(test.expectedDuration, options.BucketLifespanDuration)
			assert.Equal(test.expectedExcludedNames, options.ExcludedBucketNames)
			assert.Equal(test.expectedDryRun, options.DryRun)
			assert.Equal(test.expectedBucketNameRegex, options.BucketNameRegexp)
		})
	}
}
