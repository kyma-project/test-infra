package main

import (
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
