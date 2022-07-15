package main

import (
	"flag"
	"github.com/kyma-project/test-infra/development/gcbuild/tags"
	"os"
	"reflect"
	"testing"
)

func TestFlags(t *testing.T) {
	testcases := []struct {
		name         string
		expectedOpts options
		expectedErr  bool
		args         []string
	}{
		{
			name: "unknown flag, fail",
			expectedOpts: options{
				configFile:   "cloudbuild.yaml",
				variantsFile: "",
				buildDir:     ".",
				logDir:       "/logs/artifacts",
				project:      "sample-project",
				tagger:       tags.Tagger{TagTemplate: `v{{ .Date }}-{{ .ShortSHA }}`},
			},
			expectedErr: true,
			args: []string{
				"--project=sample-project",
				"--unknown-flag=asdasd",
			},
		},
		{
			name:        "parsed config, pass",
			expectedErr: false,
			expectedOpts: options{
				configFile:   "cloud.yaml",
				variantsFile: "var.yaml",
				buildDir:     "prow/build",
				logDir:       "prow/logs",
				project:      "sample-project",
				tagger:       tags.Tagger{TagTemplate: `{{ .CommitSHA }}`},
				devRegistry:  "eu.gcr.io/dev-registry",
				silent:       true,
			},
			args: []string{
				"--config-file=cloud.yaml",
				"--variants-file=var.yaml",
				"--project=sample-project",
				"--build-dir=prow/build",
				"--tag-template={{ .CommitSHA }}",
				"--log-dir=prow/logs",
				"--dev-registry=eu.gcr.io/dev-registry",
				"--silent",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			o := options{}
			o.gatherOptions(fs)
			if err := fs.Parse(tc.args); err != nil && !tc.expectedErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if !reflect.DeepEqual(o, tc.expectedOpts) {
				t.Errorf("%v != %v", o, tc.expectedOpts)
			}
		})
	}
}

func Test_parseVariable(t *testing.T) {
	tc := []struct {
		name     string
		expected string
		test     string
	}{
		{
			name:     "key -> _KEY=val",
			expected: "_KEY=val",
			test:     "key",
		},
		{
			name:     "_KEY -> _KEY=val",
			expected: "_KEY=val",
			test:     "_KEY",
		},
		{
			name:     "_key -> _KEY=val",
			expected: "_KEY=val",
			test:     "_key",
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			got := parseVariable(c.test, "val")
			if got != c.expected {
				t.Errorf("%s != %s", got, c.expected)
			}
		})
	}
}
