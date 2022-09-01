package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	// Registry is URL where clean build should land.
	Registry string `yaml:"directory"`
	// DevRegistry is Registry URL where development/dirty images should land.
	// If not set then the Registry field is used.
	// This field is only valid when running in CI (CI env variable is set to `true`)
	DevRegistry string `yaml:"dev-directory"`
	// Cache options that are directly related to kaniko flags
	Cache CacheConfig `yaml:"cache"`
	// TagTemplate is go-template field that defines the format of the $_TAG substitution.
	// See tags.Tag struct for more information and available fields
	TagTemplate string `yaml:"tag-template"`
	// LogFormat defines the format kaniko logs are projected.
	// Supported formats are 'color', 'text' and 'json'. Default: 'color'
	LogFormat string `yaml:"log-format"`
	// Set this option to strip timestamps out of the built image and make it Reproducible.
	Reproducible bool `yaml:"reproducible"`
}

type CacheConfig struct {
	// Enabled sets if kaniko cache is enabled or not
	Enabled bool `yaml:"enabled"`
	// CacheRunLayers sets if kaniko should cache run layers
	CacheRunLayers bool `yaml:"cache-run-layers"`
	// CacheCopyLayers sets if kaniko should cache copy layers
	CacheCopyLayers bool `yaml:"cache-copy-layers"`
	// Remote Docker directory used for cache
	CacheRepo string `yaml:"cache-repo"`
}

// ParseConfig parses yaml configuration into Config
func (c *Config) ParseConfig(f []byte) error {
	return yaml.Unmarshal(f, c)
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
