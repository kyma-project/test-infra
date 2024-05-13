package actions

import (
	"fmt"
	"os"
)

// SetOutput sets the github actions output
// It get output file name from GITHUB_OUTPUT env variable
// and writes the key=value pair to the file in the format required by github.
func SetOutput(key string, value string) error {
	// Get file path from github variable
	filePath := os.Getenv("GITHUB_OUTPUT")
	if filePath == "" {
		return fmt.Errorf("GITHUB_OUTPUT environment variable is not set, set it to valid file path")
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open output file: %s", err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	if err != nil {
		return fmt.Errorf("cannot write to output file: %s", err)
	}

	return nil
}
