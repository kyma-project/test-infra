package pathsfilter

import (
	"fmt"

	"github.com/kyma-project/test-infra/pkg/configloader"
	"go.uber.org/zap"
)

// Service is the main application service that implements the filtering logic.
type Service struct {
	log           *zap.SugaredLogger
	filesProvider ChangedFilesProvider
	resultWriter  ResultWriter
	definitions   configloader.JobDefinitions
}

// NewService creates a new instance of the application service.
func NewService(log *zap.SugaredLogger, provider ChangedFilesProvider, writer ResultWriter, definitions configloader.JobDefinitions) *Service {
	return &Service{
		log:           log,
		filesProvider: provider,
		resultWriter:  writer,
		definitions:   definitions,
	}
}

// Run executes the main logic of the paths filter.
func (s *Service) Run(eventName, targetBranch, base, head string, setOutput bool) error {
	s.log.Infow("Fetching changed files", "base", base, "head", head)
	changedFiles, err := s.filesProvider.GetChangedFiles(base, head)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	var filesForFilter []ChangedFile
	for _, f := range changedFiles {
		filesForFilter = append(filesForFilter, ChangedFile{
			Path:   f.Path,
			Status: f.Status,
		})
	}

	s.log.Infow("Found changed files", "count", len(filesForFilter))

	s.log.Infow("Applying filters...")
	filterProcessor := NewProcessor(s.definitions, s.log)
	filterResult := filterProcessor.Process(eventName, targetBranch, filesForFilter)
	s.log.Infow("Found matching filters", "count", len(filterResult.MatchedJobKeys))

	if setOutput {
		s.log.Infow("Setting outputs for GitHub Actions")
		if err := s.resultWriter.Write(filterResult); err != nil {
			return fmt.Errorf("failed to set action outputs: %w", err)
		}
	}

	return nil
}
