package configloader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EventFilters maps an event type (like "push") to a list of allowed branch names.
type EventFilters map[string][]string

// BranchFilters contains event-specific branch filtering rules.
type BranchFilters struct {
	Events EventFilters `yaml:"events"`
}

// JobDefinition represents the full configuration for a single job.
type JobDefinition struct {
	FileFilters   []string      `yaml:"file-filters"`
	BranchFilters BranchFilters `yaml:"branch-filters"`
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

	return defs, nil
}
