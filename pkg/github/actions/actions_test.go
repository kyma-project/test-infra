package actions

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
)

func TestSetOutput(t *testing.T) {
	tc := []struct {
		name             string
		predefinedOutput string
		output           string
		key              string
		value            string
		expectErr        bool
	}{
		{
			name:             "set output with empty output",
			predefinedOutput: "",
			key:              "test",
			value:            "test",
			output: `test=test
`,
		},
		{
			name: "append new output to exisitng one",
			predefinedOutput: `some=output
`,
			key:   "test",
			value: "test",
			output: `some=output
test=test
`,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			// Arrange
			// Create output file
			tempDir := t.TempDir()
			outputFilePath := fmt.Sprintf("%s/out_file", tempDir)
			err := os.WriteFile(outputFilePath, []byte(c.predefinedOutput), fs.ModePerm)
			if err != nil {
				t.Errorf("failed to write output file with predefined output: %s", err)
			}
			// Set GITHUB_OUTPUT env var
			os.Setenv("GITHUB_OUTPUT", outputFilePath)

			// Act
			err = SetOutput(c.key, c.value)
			if err != nil && !c.expectErr {
				t.Errorf("got error when not expected: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("expected error, but not got any")
			}

			// Assert
			// Read output file
			data, err := os.ReadFile(outputFilePath)
			if err != nil {
				t.Errorf("failed to read output file: %s", err)
			}
			if string(data) != c.output {
				t.Errorf("SetOutput(): Got %s, when expected: %s", data, c.output)
			}
		})
	}
}
