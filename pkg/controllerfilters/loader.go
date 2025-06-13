package controllerfilters

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EventRules holds the branch filtering rules for a single event.
type EventRules struct {
	Paths    []string `yaml:"paths"`
	Branches []string `yaml:"branches"`
}

// OnDefinition maps an event name (e.g., "pull_request_target") to its specific rules.
type OnDefinition map[string]EventRules

// JobDefinition represents the full configuration for a single job.
type JobDefinition struct {
	On OnDefinition `yaml:"on"`
}

// JobDefinitions is the top-level structure of the YAML file.
type JobDefinitions map[string]JobDefinition

// Load reads and parses a YAML filter file from the given path.
func Load(filePath string) (JobDefinitions, error) {
	var defs JobDefinitions
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("cannot parse YAML file %s: %w", filePath, err)
	}

	jobDefs := make(JobDefinitions)
	for key, value := range defs {
		if key[0] != '.' {
			jobDefs[key] = value
		}
	}

	return jobDefs, nil
}
