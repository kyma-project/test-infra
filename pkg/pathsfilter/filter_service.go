package pathsfilter

import (
	"fmt"

	"go.uber.org/zap"
)

// FilterService is the main application service that implements the filtering logic.
type FilterService struct {
	jobMatcher    JobMatcher
	filesProvider ChangedFilesProvider
	resultWriter  ResultWriter
	log           *zap.SugaredLogger
}

// NewFilterService creates a new instance of the FilterService.
func NewFilterService(
	matcher JobMatcher,
	provider ChangedFilesProvider,
	writer ResultWriter,
	log *zap.SugaredLogger,
) *FilterService {
	return &FilterService{
		jobMatcher:    matcher,
		filesProvider: provider,
		resultWriter:  writer,
		log:           log,
	}
}

// Run executes the main logic of the paths filter.
func (s *FilterService) Run(eventName string, targetBranchName string, base string, head string) error {
	s.log.Infow("Fetching changed files", "base", base, "head", head)
	changedFiles, err := s.filesProvider.GetChangedFiles(base, head)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	s.log.Infow("Found changed files", "count", len(changedFiles))
	s.log.Infow("Applying filters...")

	jobsFilterResults := s.jobMatcher.MatchJobs(eventName, targetBranchName, changedFiles)

	s.log.Infow("Found matching filters", "count", len(jobsFilterResults))
	s.log.Infow("Setting outputs for GitHub Actions")

	if err := s.resultWriter.Write(jobsFilterResults); err != nil {
		return fmt.Errorf("failed to set action outputs: %w", err)
	}

	return nil
}
