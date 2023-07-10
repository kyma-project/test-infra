package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

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

type CloudBuild struct {
	Steps         []Step            `yaml:"steps"`
	Substitutions map[string]string `yaml:"substitutions"`
	Images        []string          `yaml:"images"`
}

type Step struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"args"`
}

func GetCloudBuild(f string, fileGetter func(string) ([]byte, error)) (*CloudBuild, error) {
	b, err := fileGetter(f)
	if err != nil {
		return nil, err
	}
	var cb CloudBuild
	if err := yaml.Unmarshal(b, &cb); err != nil {
		return nil, fmt.Errorf("cloudbuild.yaml parse error: %w", err)
	}
	return &cb, nil
}

type Variants map[string]map[string]string

// GetVariants fetches variants from provided file.
// If variant flag is used, it fetches the requested variant.
func GetVariants(variant string, f string, fileGetter func(string) ([]byte, error)) (Variants, error) {
	var v Variants
	b, err := fileGetter(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// variant file not found, skipping
		return nil, nil
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	if variant != "" {
		va, ok := v[variant]
		if !ok {
			return nil, fmt.Errorf("requested variant '%s', but it's not present in variants.yaml file", variant)
		}
		return Variants{variant: va}, nil
	}
	return v, nil
}
