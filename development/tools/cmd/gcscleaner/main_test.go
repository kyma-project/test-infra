package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func cleanFlags() {
	argProjectName = ""
	argBucketLifespanDuration = bucketLifespanDurationDefault
	argExcludedBucketNames = ""
}

func TestConfigRead(t *testing.T) {

	const projectArgTag = "-project"
	const excludedBucketsArgTag = "-excludedBuckets"
	const durationArgTag = "-duration"

	tests := []struct {
		name                  string
		args                  []string
		expectedErr           error
		expectedProjectName   string
		expectedDuration      time.Duration
		expectedExcludedNames []string
	}{
		{
			name:                "just project - pass",
			args:                []string{"cmd",
				projectArgTag, "test"},
			expectedProjectName: "test",
			expectedDuration:    2 * time.Hour,
		},
		{
			name: "project, duration - pass",
			args: []string{"cmd",
				projectArgTag, "test2",
				durationArgTag, "1m"},
			expectedProjectName: "test2",
			expectedDuration:    time.Minute,
		},
		{
			name: "project, duration, excludedBuckets - pass",
			args: []string{"cmd",
				projectArgTag, "test3",
				durationArgTag, "1m",
				excludedBucketsArgTag, "test4,test5,test6"},
			expectedProjectName:   "test3",
			expectedDuration:      time.Minute,
			expectedExcludedNames: []string{"test4", "test5", "test6"},
		},
		{
			name:        "no project-name - err",
			args:        []string{"cmd"},
			expectedErr: ErrInvalidProjectName,
		},
		{
			name: "duration parsing err",
			args: []string{"cmd",
				projectArgTag, "test3",
				durationArgTag, "?"},
			expectedErr: ErrInvalidDuration,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cleanFlags()
			os.Args = test.args

			cfg, err := readConfig()
			assert := assert.New(t)
			assert.Equal(test.expectedErr, err)
			assert.Equal(test.expectedProjectName, cfg.ProjectName)
			assert.Equal(test.expectedDuration, cfg.BucketLifespanDuration)
			assert.Equal(test.expectedExcludedNames, cfg.ExcludedBucketNames)
		})
	}
}
