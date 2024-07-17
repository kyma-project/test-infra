package main

import (
	"os"
	"testing"
)

func TestGetTarget(t *testing.T) {
	tests := []struct {
		source     string
		targetRepo string
		targetTag  string
		expected   string
		shouldErr  bool
	}{
		{
			source:     "cypress/included:9.5.0",
			targetRepo: "external/prod/",
			targetTag:  "latest",
			expected:   "external/prod/cypress/included:latest",
			shouldErr:  false,
		},
		{
			source:     "included:9.5.0",
			targetRepo: "external/prod/",
			targetTag:  "latest",
			expected:   "external/prod/included:latest",
			shouldErr:  false,
		},
		{
			source:     "cypress/included@sha256:abcdef",
			targetRepo: "external/prod/",
			targetTag:  "latest",
			expected:   "external/prod/cypress/included:latest",
			shouldErr:  false,
		},
		{
			source:     "included@sha256:abcdef",
			targetRepo: "external/prod/",
			targetTag:  "latest",
			expected:   "external/prod/included:latest",
			shouldErr:  false,
		},
		{
			source:     "included@sha256:abcdef",
			targetRepo: "external/prod/",
			targetTag:  "",
			expected:   "",
			shouldErr:  true,
		},
		{
			source:     "cypress/included:9.5.0",
			targetRepo: "external/prod/",
			targetTag:  "",
			expected:   "external/prod/cypress/included:9.5.0",
			shouldErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			result, err := getTarget(test.source, test.targetRepo, test.targetTag)
			if (err != nil) != test.shouldErr {
				t.Errorf("expected error: %v, got: %v", test.shouldErr, err)
			}
			if result != test.expected {
				t.Errorf("expected: %s, got: %s", test.expected, result)
			}
		})
	}
}

func TestParseImagesFile(t *testing.T) {
	validYAML := `
targetRepoPrefix: "external/prod/"
images:
  - source: "cypress/included:9.5.0"
    target: "external/prod/cypress/included:latest"
`

	missingTargetRepoPrefixYAML := `
images:
  - source: "cypress/included:9.5.0"
    target: "external/prod/cypress/included:latest"
`

	tests := []struct {
		name        string
		content     string
		shouldErr   bool
		expectedErr string
	}{
		{
			name:      "valid YAML",
			content:   validYAML,
			shouldErr: false,
		},
		{
			name:        "missing targetRepoPrefix",
			content:     missingTargetRepoPrefixYAML,
			shouldErr:   true,
			expectedErr: "targetRepoPrefix can not be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "test*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(file.Name())

			if _, err := file.Write([]byte(test.content)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := file.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			_, err = parseImagesFile(file.Name())
			if (err != nil) != test.shouldErr {
				t.Errorf("expected error: %v, got: %v", test.shouldErr, err)
			}
			if err != nil && test.shouldErr && err.Error() != test.expectedErr {
				t.Errorf("expected error message: %s, got: %s", test.expectedErr, err.Error())
			}
		})
	}
}
