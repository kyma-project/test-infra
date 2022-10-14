package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"testing/fstest"
)

func Test_gatherDestinations(t *testing.T) {
	tc := []struct {
		name      string
		directory string
		repos     []string
		tags      []string
		expected  []string
	}{
		{
			name:      "test-image",
			repos:     []string{"dev.kyma.io", "dev2.kyma.io"},
			directory: "subdirectory",
			tags: []string{
				"20222002-abcd1234",
				"latest",
				"cookie",
			},
			expected: []string{
				"dev.kyma.io/subdirectory/test-image:20222002-abcd1234",
				"dev2.kyma.io/subdirectory/test-image:20222002-abcd1234",
				"dev.kyma.io/subdirectory/test-image:latest",
				"dev2.kyma.io/subdirectory/test-image:latest",
				"dev.kyma.io/subdirectory/test-image:cookie",
				"dev2.kyma.io/subdirectory/test-image:cookie",
			},
		},
		{
			name:     "test-no-directory",
			repos:    []string{"kyma.dev"},
			tags:     []string{"latest"},
			expected: []string{"kyma.dev/test-no-directory:latest"},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			got := gatherDestinations(c.repos, c.directory, c.name, c.tags)
			if len(c.expected) != len(got) {
				t.Errorf("result length mismatch. wanted %v, got %v", len(c.expected), len(got))
			}
			if !reflect.DeepEqual(c.expected, got) {
				t.Errorf("%s != %s", got, c.expected)
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
			name:     "key -> key=val",
			expected: "key=val",
			test:     "key",
		},
		{
			name:     "_KEY -> _KEY=val",
			expected: "_KEY=val",
			test:     "_KEY",
		},
		{
			name:     "_key -> _key=val",
			expected: "_key=val",
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
			name:      "parsed config",
			expectErr: false,
			opts: options{
				directory:  "kyma.dev",
				context:    "directory/",
				name:       "test-image",
				dockerfile: "Dockerfile",
			},
		},
		{
			name:      "context missing",
			expectErr: true,
			opts: options{
				directory:  "kyma.dev",
				name:       "test-image",
				dockerfile: "Dockerfile",
			},
		},
		{
			name:      "name missing",
			expectErr: true,
			opts: options{
				directory:  "kyma.dev",
				context:    "directory/",
				dockerfile: "Dockerfile",
			},
		},
		{
			name:      "dockerfile missing",
			expectErr: true,
			opts: options{
				directory: "kyma.dev",
				context:   "directory/",
				name:      "test-image",
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
				context:    ".",
				configPath: "/config/image-builder-config.yaml",
				dockerfile: "Dockerfile",
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
				name:       "test-image",
				directory:  "subdirectory",
				tags:       []string{"latest", "cookie"},
				context:    "prow/build",
				configPath: "config.yaml",
				dockerfile: "Dockerfile",
				logDir:     "prow/logs",
				silent:     true,
			},
			args: []string{
				"--config=config.yaml",
				"--dockerfile=Dockerfile",
				"--directory=subdirectory",
				"--name=test-image",
				"--tag=latest",
				"--tag=cookie",
				"--context=prow/build",
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

func Test_gatTags(t *testing.T) {
	tc := []struct {
		name           string
		pr             string
		sha            string
		tagTemplate    string
		env            map[string]string
		additionalTags []string
		expectErr      bool
		expectResult   []string
	}{
		{
			name:         "pr variable is present",
			pr:           "1234",
			expectResult: []string{"PR-1234"},
		},
		{
			name:      "sha is empty",
			expectErr: true,
		},
		{
			name:        "bad tagTemplate",
			expectErr:   true,
			sha:         "abcd1234",
			tagTemplate: `v{{ .ASD }}`,
		},
		{
			name:           "custom template, additional fields, env variable",
			expectErr:      false,
			sha:            "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			tagTemplate:    `{{ .ShortSHA }}`,
			env:            map[string]string{"CUSTOM_ENV": "customEnvValue"},
			additionalTags: []string{"latest", "cookie", `{{ .CommitSHA }}`, `{{ .Env "CUSTOM_ENV" }}`},
			expectResult:   []string{"latest", "cookie", "da39a3ee5e6b4b0d3255bfef95601890afd80709", "customEnvValue", "da39a3ee"},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			got, err := getTags(c.pr, c.sha, append(c.additionalTags, c.tagTemplate))
			if err != nil && !c.expectErr {
				t.Errorf("got error but didn't want to: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't get error but wanted to")
			}
			if !reflect.DeepEqual(c.expectResult, got) {
				t.Errorf("%v != %v", got, c.expectResult)
			}
		})
	}
}

func Test_loadEnv(t *testing.T) {
	// static value that should not be overridden
	t.Setenv("key3", "static-value")
	vfs := fstest.MapFS{
		".env": &fstest.MapFile{Data: []byte("KEY=VAL\nkey2=val2\nkey3=val3\nkey4=val4=asf"), Mode: 0666},
	}
	expected := map[string]string{
		"KEY":  "VAL",
		"key2": "val2",
		"key3": "static-value",
		"key4": "val4=asf",
	}
	_, err := loadEnv(vfs, ".env")
	if err != nil {
		t.Errorf("%v", err)
	}

	for k, v := range expected {
		got := os.Getenv(k)
		if got != v {
			t.Errorf("%v != %v", got, v)
		}
		os.Unsetenv(k)
	}
}
