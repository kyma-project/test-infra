package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/kyma-project/test-infra/pkg/tags"
)

func Test_ParseConfig(t *testing.T) {
	tc := []struct {
		name           string
		config         string
		expectedConfig Config
		expectErr      bool
	}{
		{
			name: "parsed full config one repo",
			config: `registry: kyma-project.io/prod-registry
dev-registry: dev.kyma-project.io/dev-registry
tag-template: v{{ .Date }}-{{ .ShortSHA }}`,
			expectedConfig: Config{
				Registry:    []string{"kyma-project.io/prod-registry"},
				DevRegistry: []string{"dev.kyma-project.io/dev-registry"},
				TagTemplate: tags.Tag{Name: "default_tag", Value: `v{{ .Date }}-{{ .ShortSHA }}`},
			},
		},
		{
			name: "parsed full config with multiple repos",
			config: `registry:
- kyma-project.io/prod-registry
- kyma-project.io/second-registry
dev-registry:
- dev.kyma-project.io/dev-registry
- dev.kyma-project.io/second-registry
tag-template: v{{ .Date }}-{{ .ShortSHA }}`,
			expectedConfig: Config{
				Registry:    []string{"kyma-project.io/prod-registry", "kyma-project.io/second-registry"},
				DevRegistry: []string{"dev.kyma-project.io/dev-registry", "dev.kyma-project.io/second-registry"},
				TagTemplate: tags.Tag{Name: "default_tag", Value: `v{{ .Date }}-{{ .ShortSHA }}`},
			},
		},
		{
			name:           "malformed yaml file, fail",
			config:         `garbage:malformed:config`,
			expectedConfig: Config{},
			expectErr:      true,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			o := Config{}
			err := o.ParseConfig([]byte(c.config))
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if !reflect.DeepEqual(o, c.expectedConfig) {
				t.Errorf("%v != %v", o, c.expectedConfig)
			}
		})
	}
}

func Test_getVariants(t *testing.T) {
	tc := []struct {
		name           string
		variant        string
		expectVariants Variants
		expectErr      bool
		variantsFile   string
	}{
		{
			name:         "valid file, variants passed",
			expectErr:    false,
			variantsFile: "variants.yaml",
			expectVariants: Variants{
				"main": map[string]string{"KUBECTL_VERSION": "1.24.4"},
				"1.23": map[string]string{"KUBECTL_VERSION": "1.23.9"},
			},
		},
		{
			name:           "variant file does not exist, pass",
			expectErr:      false,
			variantsFile:   "",
			expectVariants: nil,
		},
		{
			name:           "other error during getting file, fail",
			expectErr:      true,
			variantsFile:   "err-deadline-exceeded",
			expectVariants: nil,
		},
		{
			name:         "get only single variant, pass",
			expectErr:    false,
			variantsFile: "variants.yaml",
			variant:      "main",
			expectVariants: Variants{
				"main": map[string]string{"KUBECTL_VERSION": "1.24.4"},
			},
		},
		{
			name:           "variant is not present in variants file, fail",
			expectErr:      true,
			variantsFile:   "variants.yaml",
			variant:        "missing-variant",
			expectVariants: nil,
		},
		{
			name:           "malformed variants file, fail",
			expectErr:      true,
			variantsFile:   "malformed-variants.yaml",
			expectVariants: nil,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			fakeFileGetter := func(f string) ([]byte, error) {
				if f == "malformed-variants.yaml" {
					return []byte("'asd':\n- malformed variant as list`"), nil
				}
				if f == "err-deadline-exceeded" {
					return nil, os.ErrDeadlineExceeded
				}
				vf := "'main':\n  KUBECTL_VERSION: \"1.24.4\"\n'1.23':\n  KUBECTL_VERSION: \"1.23.9\""
				if f == "variants.yaml" {
					return []byte(vf), nil
				}
				return nil, os.ErrNotExist
			}

			v, err := GetVariants(c.variant, c.variantsFile, fakeFileGetter)
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't catch error, but wanted to")
			}
			if !reflect.DeepEqual(v, c.expectVariants) {
				t.Errorf("%v != %v", v, c.expectVariants)
			}
		})
	}
}

func TestLoadGitStateConfigFromEnv(t *testing.T) {
	tc := []struct {
		name        string
		options     options
		env         map[string]string
		gitState    GitStateConfig
		expectError bool
	}{
		{
			name: "Load config from ProwJobEnv for ADO presubmit",
			options: options{
				buildInADO: true,
			},
			env: map[string]string{
				"REPO_NAME":     "test-repo",
				"REPO_OWNER":    "test-owner",
				"JOB_TYPE":      "presubmit",
				"PULL_NUMBER":   "1234",
				"PULL_BASE_SHA": "art654",
				"PULL_PULL_SHA": "qwe456",
			},
			gitState: GitStateConfig{
				RepositoryName:    "test-repo",
				RepositoryOwner:   "test-owner",
				JobType:           "presubmit",
				PullRequestNumber: "1234",
				BaseCommitSHA:     "art654",
				PullHeadCommitSHA: "qwe456",
			},
		},
		{
			name: "Invalid job type value in prowjob env",
			options: options{
				buildInADO: true,
			},
			env: map[string]string{
				"REPO_NAME":     "test-repo",
				"REPO_OWNER":    "test-owner",
				"JOB_TYPE":      "periodic",
				"PULL_NUMBER":   "1234",
				"PULL_BASE_SHA": "art654",
				"PULL_PULL_SHA": "qwe456",
			},
			expectError: true,
		},
		{
			name: "Missing repo name value in prowjob env",
			options: options{
				buildInADO: true,
			},
			env: map[string]string{
				"REPO_OWNER":    "test-owner",
				"JOB_TYPE":      "periodic",
				"PULL_NUMBER":   "1234",
				"PULL_BASE_SHA": "art654",
				"PULL_PULL_SHA": "qwe456",
			},
			expectError: true,
		},
		{
			name: "Load data from event payload for github pull_request_target",
			options: options{
				runInActions: true,
			},
			env: map[string]string{
				"GITHUB_EVENT_PATH": "./test_fixture/pull_request_target_reopened.json",
				"GITHUB_EVENT_NAME": "pull_request_target",
			},
			gitState: GitStateConfig{
				RepositoryName:    "test-infra",
				RepositoryOwner:   "KacperMalachowski",
				JobType:           "presubmit",
				PullRequestNumber: "5",
				BaseCommitSHA:     "4b91c74a2aa9aeeb4a265cf1ffe2dd54812b4124",
				PullHeadCommitSHA: "df7ebcecce2ec1299b5e8ccb22482f901d205abc",
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			// Prepare env vars
			for key, value := range c.env {
				err := os.Setenv(key, value)
				if err != nil {
					t.Errorf("failed to set env variable %s with value %s", key, value)
				}
			}

			// Load git state
			state, err := LoadGitStateConfigFromEnv(c.options)
			if err != nil && !c.expectError {
				t.Errorf("unexpected error occured %s", err)
			}
			if err == nil && c.expectError {
				t.Error("expected error, but not occured")
			}

			if !reflect.DeepEqual(state, c.gitState) {
				t.Errorf("LoadGitStateConfigFromEnv(): Got %v, but expected %v", state, c.gitState)
			}
		})
	}
}
