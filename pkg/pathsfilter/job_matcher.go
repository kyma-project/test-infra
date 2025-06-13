package pathsfilter

import (
	"github.com/kyma-project/test-infra/pkg/controllerfilters"
	"github.com/kyma-project/test-infra/pkg/matcher"
	"go.uber.org/zap"
)

// JobFiltersResult holds the outcome of a filtering operation.
type JobFiltersResult struct {
	TriggeredJobKeys []string
	JobTriggers      map[string]bool
}

// Job represents a single job with all its filtering rules.
type Job struct {
	Name    string
	OnRules controllerfilters.OnDefinition
	log     *zap.SugaredLogger
}

// shouldRun determines if a job should be triggered based on event, branch, and file changes.
func (j *Job) shouldRun(eventName string, targetBranchName string, changedFiles []string) bool {
	eventRules, eventDefined := j.OnRules[eventName]
	if !eventDefined {
		j.log.Debugw("Job skipped, event not defined in config", "job", j.Name, "event", eventName)

		return false
	}

	branchMatch := j.isBranchAllowed(eventRules.Branches, targetBranchName)
	fileMatch := j.hasMatchingFileChanges(eventRules.Paths, changedFiles)

	j.log.Debugw("Condition evaluation for job",
		"job", j.Name,
		"branch_match", branchMatch,
		"file_match", fileMatch,
	)

	return branchMatch && fileMatch
}

// isBranchAllowed checks if the target branch is in the list of allowed branches for the current event.
func (j *Job) isBranchAllowed(allowedBranches []string, targetBranchName string) bool {
	if len(allowedBranches) == 0 {
		return true
	}

	for _, allowedBranchName := range allowedBranches {
		if allowedBranchName == targetBranchName {
			return true
		}
	}

	return false
}

// hasMatchingFileChanges checks if any changed files match the job's file path patterns.
func (j *Job) hasMatchingFileChanges(pathPatterns []string, changedFiles []string) bool {
	if len(pathPatterns) == 0 {
		return true
	}

	for _, pattern := range pathPatterns {
		for _, filePath := range changedFiles {
			if ok, _ := matcher.Match(pattern, filePath); ok {
				return true
			}
		}
	}

	return false
}

// JobMatcher encapsulates the filtering logic for a set of job definitions.
type JobMatcher struct {
	jobs []Job
	log  *zap.SugaredLogger
}

// NewJobMatcher creates a new instance of JobMatcher.
func NewJobMatcher(definitions controllerfilters.JobDefinitions, log *zap.SugaredLogger) JobMatcher {
	var jobs []Job
	for name, def := range definitions {
		jobs = append(jobs, Job{
			Name:    name,
			OnRules: def.On,
			log:     log,
		})
	}

	return JobMatcher{jobs: jobs, log: log}
}

// MatchJobs is the primary method that applies all filters to a list of changed files.
func (p *JobMatcher) MatchJobs(eventName string, targetBranchName string, changedFiles []string) JobFiltersResult {
	result := JobFiltersResult{
		TriggeredJobKeys: []string{},
		JobTriggers:      make(map[string]bool),
	}

	for _, job := range p.jobs {
		shouldRun := job.shouldRun(eventName, targetBranchName, changedFiles)
		result.JobTriggers[job.Name] = shouldRun
		if shouldRun {
			result.TriggeredJobKeys = append(result.TriggeredJobKeys, job.Name)
		}
	}

	return result
}
