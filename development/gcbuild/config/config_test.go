package config

import (
	"os"
	"reflect"
	"testing"
)

func Test_ParseConfig(t *testing.T) {
	tc := []struct {
		name           string
		config         string
		expectedConfig Config
		expectErr      bool
	}{
		{
			name: "parsed full config",
			config: `project: sample-project
devRegistry: dev.kyma-project.io/dev-registry
stagingBucket: gs://staging-bucket
logsBucket: gs://logs-bucket
tagTemplate: v{{ .Date }}-{{ .ShortSHA }}`,
			expectedConfig: Config{
				Project:       "sample-project",
				DevRegistry:   "dev.kyma-project.io/dev-registry",
				StagingBucket: "gs://staging-bucket",
				LogsBucket:    "gs://logs-bucket",
				TagTemplate:   `v{{ .Date }}-{{ .ShortSHA }}`,
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

func Test_getCloudbuild(t *testing.T) {
	tc := []struct {
		name             string
		expectErr        bool
		expectCloudbuild *CloudBuild
		cloudbuildFile   string
	}{
		{
			name:           "missing cloudbuild.yaml, fail",
			cloudbuildFile: "missing-cloudbuild.yaml",
			expectErr:      true,
		},
		{
			name:           "malformed cloudbuild, fail",
			cloudbuildFile: "malformed-cloudbuild.yaml",
			expectErr:      true,
		},
		{
			name:           "valid cloudbuild, pass",
			cloudbuildFile: "cloudbuild.yaml",
			expectErr:      false,
			expectCloudbuild: &CloudBuild{
				Steps: []Step{
					{
						Name: "gcr.io/cloud-builders/docker",
						Args: []string{
							"build",
							"--tag=$_REPOSITORY/test:$_TAG",
							"--tag=$_REPOSITORY/test:latest",
							".",
						},
					},
				},
				Substitutions: map[string]string{"_REPOSITORY": "kyma.dev"},
				Images:        []string{"$_REPOSITORY/test:$_TAG", "$_REPOSITORY/test:latest"},
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			fakeFileGetter := func(f string) ([]byte, error) {
				if f == "malformed-cloudbuild.yaml" {
					return []byte(`steps:
  id: build-image
  name: 'gcr.io/cloud-builders/docker'
  args:
    - build
    - --tag=$_REPOSITORY/test:$_TAG
    - --tag=$_REPOSITORY/test:latest
    - .
  dir: .
substitutions:
  _REPOSITORY: kyma.dev
images:
  - $_REPOSITORY/test:$_TAG
  - $_REPOSITORY/test:latest`), nil
				}
				if f == "cloudbuild.yaml" {
					return []byte(`steps:
  - id: build-image
    name: 'gcr.io/cloud-builders/docker'
    args:
      - build
      - --tag=$_REPOSITORY/test:$_TAG
      - --tag=$_REPOSITORY/test:latest
      - .
    dir: .
substitutions:
  _REPOSITORY: kyma.dev
images:
  - $_REPOSITORY/test:$_TAG
  - $_REPOSITORY/test:latest`), nil
				}
				return nil, os.ErrNotExist
			}

			cb, err := GetCloudBuild(c.cloudbuildFile, fakeFileGetter)
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't catch error, but wanted to")
			}
			if !reflect.DeepEqual(cb, c.expectCloudbuild) {
				t.Errorf("%v != %v", cb, c.expectCloudbuild)
			}
		})
	}
}
