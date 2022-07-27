package main

import "gopkg.in/yaml.v3"

type Config struct {
	DevRegistry   string `yaml:"devRegistry"`
	Project       string `yaml:"project"`
	StagingBucket string `yaml:"stagingBucket,omitempty"`
	LogsBucket    string `yaml:"logsBucket,omitempty"`
}

func (c *Config) ParseConfig(f []byte) error {
	return yaml.Unmarshal(f, c)
}
