package configloader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Definition represents the structure of a filter file.
// The key is the filter name and the value is a slice of glob patterns.
type Definition map[string][]string

// Load reads and parses a YAML filter file from the given path.
func Load(filePath string) (Definition, error) {
	var defs Definition
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", filePath, err)
	}
	if err := yaml.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("cannot parse YAML file %s: %w", filePath, err)
	}

	return defs, nil
}
