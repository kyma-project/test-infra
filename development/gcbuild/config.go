package main

import "gopkg.in/yaml.v3"

type Config struct {
	// DevRegistry is Registry URL where development/dirty images should land.
	// If not set then the default registry is used.
	// This field is only valid when running in CI (CI env variable is set to `true`)
	DevRegistry string `yaml:"devRegistry"`
	// Project is GCP project name where build jobs will run
	// This field is required
	Project string `yaml:"project"`
	// StagingBucket is the name of the Google Cloud Storage bucket, where the source will be pushed beforehand.
	// If not set, rely on Google Cloud Build
	StagingBucket string `yaml:"stagingBucket,omitempty"`
	// LogsBucket is the name to the Google Cloud Storage bucket, where the logs will be pushed after build finishes.
	// If not set, rely on Google Cloud Build
	LogsBucket string `yaml:"logsBucket,omitempty"`
	// TagTemplate is go-template field that defines the format of the $_TAG substitution.
	// See tags.Tag struct for more information and available fields
	TagTemplate string `yaml:"tagTemplate"`
}

// ParseConfig parses yaml configuration into Config
func (c *Config) ParseConfig(f []byte) error {
	return yaml.Unmarshal(f, c)
}
