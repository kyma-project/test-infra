package pathsfilter

import (
	"github.com/kyma-project/test-infra/pkg/configloader"
	"github.com/kyma-project/test-infra/pkg/matcher"
	"go.uber.org/zap"
)

// Result holds the outcome of a filtering operation.
type Result struct {
	MatchedJobKeys          []string
	IndividualJobRunResults map[string]bool
}

// Job represents a single job with all its filtering rules.
type Job struct {
	Name          string
	FilePatterns  []string
	BranchFilters configloader.BranchFilters
	log           *zap.SugaredLogger
}

// shouldRun determines if a job should be triggered based on event, branch, and file changes.
func (j *Job) shouldRun(eventName string, targetBranch string, changedFiles []ChangedFile) bool { // CHANGE: Uses the local ChangedFile type
	if !j.matchesBranch(eventName, targetBranch) {
		j.log.Debugw("Job skipped due to branch/event filter mismatch", "job", j.Name, "event", eventName, "branch", targetBranch)
		return false
	}

	if !j.matchesFiles(changedFiles) {
		j.log.Debugw("Job skipped due to file filter mismatch", "job", j.Name)
		return false
	}

	j.log.Debugw("Job conditions met, will be triggered", "job", j.Name)
	return true
}

// matchesBranch checks if the current event and branch match the job's branch filters.
func (j *Job) matchesBranch(eventName string, targetBranch string) bool {
	if j.BranchFilters.Events == nil {
		return true
	}

	allowedBranches, eventDefined := j.BranchFilters.Events[eventName]
	if !eventDefined {
		return false
	}

	for _, allowedBranch := range allowedBranches {
		if allowedBranch == targetBranch {
			return true
		}
	}

	return false
}

// matchesFiles checks if any changed files match the job's file patterns.
func (j *Job) matchesFiles(changedFiles []ChangedFile) bool { // CHANGE: Uses the local ChangedFile type
	if len(j.FilePatterns) == 0 {
		return true
	}

	for _, pattern := range j.FilePatterns {
		for _, file := range changedFiles {
			// Original file used matcher.Match(pattern, file.Path)
			if ok, _ := matcher.Match(pattern, file.Path); ok {
				return true
			}
		}
	}
	return false
}

// Processor encapsulates the filtering logic for a set of job definitions.
type Processor struct {
	jobs []Job
	log  *zap.SugaredLogger
}

// NewProcessor creates a new filter processor from job definitions.
func NewProcessor(definitions configloader.JobDefinitions, log *zap.SugaredLogger) *Processor {
	var jobs []Job
	for name, def := range definitions {
		jobs = append(jobs, Job{
			Name:          name,
			FilePatterns:  def.FileFilters,
			BranchFilters: def.BranchFilters,
			log:           log,
		})
	}

	return &Processor{jobs: jobs, log: log}
}

// Process is the primary method that applies all filters to a list of changed files.
func (p *Processor) Process(eventName string, targetBranch string, changedFiles []ChangedFile) Result { // CHANGE: Uses the local ChangedFile type
	result := Result{
		MatchedJobKeys:          []string{},
		IndividualJobRunResults: make(map[string]bool),
	}

	for _, job := range p.jobs {
		shouldRun := job.shouldRun(eventName, targetBranch, changedFiles)
		result.IndividualJobRunResults[job.Name] = shouldRun
		if shouldRun {
			result.MatchedJobKeys = append(result.MatchedJobKeys, job.Name)
		}
	}

	return result
}
