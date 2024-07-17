package main

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMainSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = Describe("getTarget", func() {
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
			expected:   "external/prod/library/included:latest",
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
			expected:   "external/prod/library/included:latest",
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
		test := test // capture range variable
		It("should handle "+test.source, func() {
			result, err := getTarget(test.source, test.targetRepo, test.targetTag)
			if test.shouldErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(test.expected))
			}
		})
	}
})

var _ = Describe("parseImagesFile", func() {
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
		test := test // capture range variable
		It("should handle "+test.name, func() {
			file, err := os.CreateTemp("", "test*.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(file.Name())

			_, err = file.Write([]byte(test.content))
			Expect(err).NotTo(HaveOccurred())
			Expect(file.Close()).To(Succeed())

			_, err = parseImagesFile(file.Name())
			if test.shouldErr {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(test.expectedErr))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		})
	}
})
