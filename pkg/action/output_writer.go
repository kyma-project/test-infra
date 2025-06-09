package actions

import (
	"encoding/json" // Import for JSON functionality
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/pkg/pathsfilter"
	"go.uber.org/zap"
)

// OutputWriter implements the pathsfilter.ResultWriter port.
type OutputWriter struct {
	log *zap.SugaredLogger
}

// NewOutputWriter creates a new output writer adapter.
func NewOutputWriter(log *zap.SugaredLogger) *OutputWriter {
	return &OutputWriter{log: log}
}

// Write marshals the filtering result into a single JSON object and writes it
func (w *OutputWriter) Write(result pathsfilter.Result) error {
	outputFilePath := os.Getenv("GITHUB_OUTPUT")
	if outputFilePath == "" {
		w.log.Warnw("GITHUB_OUTPUT environment variable not set. Skipping writing outputs.")
		return nil
	}

	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open GITHUB_OUTPUT file %s: %w", outputFilePath, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			w.log.Errorw("Could not close GITHUB_OUTPUT file", "path", outputFilePath, "error", err)
		}
	}(file)

	w.log.Infow("Marshalling results to JSON...")
	jsonOutput, err := json.Marshal(result.IndividualJobRunResults)
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	outputName := "changes-json"
	outputValue := string(jsonOutput)

	w.log.Infow("Writing JSON output to file.", "name", outputName, "value", outputValue)
	if err := w.set(file, outputName, outputValue); err != nil {
		return err
	}

	w.log.Infow("Finished writing outputs.")

	return nil
}

// set writes a single key-value pair to the GitHub Actions output file.
func (w *OutputWriter) set(file *os.File, key, value string) error {
	w.log.Debugw("Setting output", "key", key, "value", value)
	if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
		return fmt.Errorf("failed to write to GITHUB_OUTPUT for key %s: %w", key, err)
	}

	return nil
}
