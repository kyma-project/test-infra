package actions

import (
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

// Write writes the results to the GITHUB_OUTPUT file.
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
			fmt.Printf("could not close GITHUB_OUTPUT file %s: %v\n", outputFilePath, err)
		}
	}(file)

	w.log.Infow("Writing individual job run results to output...")
	for key, shouldRun := range result.IndividualJobRunResults {
		if err := w.set(file, key, fmt.Sprintf("%t", shouldRun)); err != nil {
			return err
		}
	}

	w.log.Infow("Finished writing outputs.")

	return nil
}

func (w *OutputWriter) set(file *os.File, key, value string) error {
	w.log.Debugw("Setting output", "key", key, "value", value)
	if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
		return fmt.Errorf("failed to write to GITHUB_OUTPUT for key %s: %w", key, err)
	}

	return nil
}
