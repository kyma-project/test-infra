package filter

import (
	"github.com/kyma-project/test-infra/pkg/configloader"
	"github.com/kyma-project/test-infra/pkg/git"
	"github.com/kyma-project/test-infra/pkg/matcher"
	"go.uber.org/zap"
)

// Result holds the outcome of a filtering operation.
type Result struct {
	MatchedFilterKeys       []string
	IndividualFilterResults map[string]bool
	MatchedFilesByFilter    map[string][]string
}

// Filter represents a single named filter with its associated glob patterns.
type Filter struct {
	Name     string
	Patterns []string
}

// findFilesMatchingPattern checks a single pattern against a list of files and returns matches.
func findFilesMatchingPattern(pattern string, files []git.ChangedFile) []string {
	var matchedFiles []string
	for _, file := range files {
		if ok, _ := matcher.Match(pattern, file.Path); ok {
			matchedFiles = append(matchedFiles, file.Path)
		}
	}

	return matchedFiles
}

// matches checks if this filter is matched by any of the changed files.
func (f *Filter) matches(changedFiles []git.ChangedFile) (bool, []string) {
	var allMatchedFiles []string
	for _, pattern := range f.Patterns {
		matchedForPattern := findFilesMatchingPattern(pattern, changedFiles)
		allMatchedFiles = append(allMatchedFiles, matchedForPattern...)
	}

	if len(allMatchedFiles) > 0 {
		return true, unique(allMatchedFiles)
	}

	return false, nil
}

// Processor encapsulates the filtering logic for a set of filter definitions.
type Processor struct {
	filters []Filter
	log     *zap.SugaredLogger
}

// NewProcessor creates a new filter processor from filter definitions.
func NewProcessor(definitions configloader.Definition, log *zap.SugaredLogger) *Processor {
	var filters []Filter
	for name, patterns := range definitions {
		filters = append(filters, Filter{Name: name, Patterns: patterns})
	}
	return &Processor{filters: filters, log: log}
}

// Process is the primary method that applies all filters to a list of changed files.
func (p *Processor) Process(changedFiles []git.ChangedFile) Result {
	result := Result{
		MatchedFilterKeys:       []string{},
		IndividualFilterResults: make(map[string]bool),
		MatchedFilesByFilter:    make(map[string][]string),
	}

	for _, f := range p.filters {
		result.IndividualFilterResults[f.Name] = false
	}

	for _, f := range p.filters {
		p.log.Debugw("Checking filter", "key", f.Name)
		if isMatch, matchedFiles := f.matches(changedFiles); isMatch {
			p.log.Debugw("Found match for filter", "key", f.Name, "files_count", len(matchedFiles))
			result.MatchedFilterKeys = append(result.MatchedFilterKeys, f.Name)
			result.IndividualFilterResults[f.Name] = true
			result.MatchedFilesByFilter[f.Name] = matchedFiles
		}
	}

	return result
}

// unique is a helper function to remove duplicates from a slice of strings.
func unique(slice []string) []string {
	keys := make(map[string]struct{})
	var list []string
	for _, entry := range slice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}

	return list
}
