package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/kyma-project/test-infra/development/pkg/tags"
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
