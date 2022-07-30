package main

import (
	"flag"
	"github.com/kyma-project/test-infra/development/gcbuild/config"
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
				cloudbuild: "cloudbuild.yaml",
				buildDir:   ".",
				logDir:     "/logs/artifacts",
			},
			expectedErr: true,
			args: []string{
				"--unknown-flag=asdasd",
			},
		},
		{
			name:        "parsed config, pass",
			expectedErr: false,
			expectedOpts: options{
				cloudbuild: "cloud.yaml",
				configPath: "config.yaml",
				buildDir:   "prow/build",
				logDir:     "prow/logs",
				silent:     true,
			},
			args: []string{
				"--config=config.yaml",
				"--cloudbuild-file=cloud.yaml",
				"--build-dir=prow/build",
				"--log-dir=prow/logs",
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

func Test_validateOptions(t *testing.T) {
	tc := []struct {
		name      string
		expectErr bool
		opts      options
	}{
		{
			name:      "project is missing",
			expectErr: true,
			opts: options{
				buildDir:   "dir/",
				cloudbuild: "cloud.yaml",
			},
		},
		{
			name:      "buildDir is missing",
			expectErr: true,
			opts: options{
				cloudbuild: "cloud.yaml",
				Config:     config.Config{Project: "sample-project"},
			},
		},
		{
			name:      "cloudbuild is missing",
			expectErr: true,
			opts: options{
				buildDir: ".",
				Config:   config.Config{Project: "sample-project"},
			},
		},
		{
			name:      "options are valid",
			expectErr: false,
			opts: options{
				buildDir:   ".",
				cloudbuild: "cloud.yaml",
				variant:    "main",
				Config:     config.Config{Project: "sample-project"},
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err := validateOptions(c.opts)
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't catch error, but wanted to")
			}
		})
	}
}

func Test_getImageNames(t *testing.T) {
	tc := []struct {
		name         string
		images       []string
		expectString string
	}{
		{
			name: "replaced strings without {}",
			images: []string{
				"$_REPOSITORY/image:$_TAG-$_VARIANT",
				"$_REPOSITORY/image:latest",
			},
			expectString: "kyma.dev/image:12345678-abcdef-main kyma.dev/image:latest",
		},
		{
			name: "replaced strings with {}",
			images: []string{
				"${_REPOSITORY}/image:${_TAG}-${_VARIANT}",
				"${_REPOSITORY}/image:latest",
			},
			expectString: "kyma.dev/image:12345678-abcdef-main kyma.dev/image:latest",
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			repo := "kyma.dev"
			tag := "12345678-abcdef"
			variant := "main"
			s := getImageNames(repo, tag, variant, c.images)
			if s != c.expectString {
				t.Errorf("%s != %s", s, c.expectString)
			}
		})
	}
}
