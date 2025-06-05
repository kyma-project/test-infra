package action

import (
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/pkg/filter"
	"go.uber.org/zap"
)

// OutputWriter implements the logic for writing to the GITHUB_OUTPUT file.
type OutputWriter struct {
	log *zap.SugaredLogger
}

// NewOutputWriter creates a new instance of OutputWriter.
func NewOutputWriter(log *zap.SugaredLogger) *OutputWriter {
	return &OutputWriter{log: log}
}

// Write processes the filter result and writes it as action outputs.
func (w *OutputWriter) Write(result filter.Result) error {
	outputFilePath := os.Getenv("GITHUB_OUTPUT")
	if outputFilePath == "" {
		w.log.Infow("GITHUB_OUTPUT environment variable not set. Skipping writing outputs.")
		return nil
	}

	file, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open GITHUB_OUTPUT file %s: %w", outputFilePath, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			w.log.Errorw("could not close the file", file.Name())
		}
	}(file)

	for key, matched := range result.IndividualFilterResults {
		if err := w.set(file, key, fmt.Sprintf("%t", matched)); err != nil {
			return err
		}
	}

	return nil
}

// set is a helper to write a single key-value pair to the output file.
func (w *OutputWriter) set(file *os.File, key, value string) error {
	w.log.Infow("Setting output", "key", key, "value", value)
	if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
		return fmt.Errorf("failed to write to GITHUB_OUTPUT for key %s: %w", key, err)
	}

	return nil
}
