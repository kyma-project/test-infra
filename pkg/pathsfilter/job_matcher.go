package pathsfilter

import (
	"github.com/kyma-project/test-infra/pkg/configloader"
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
	Name         string
	PathPatterns []string
	OnRules      configloader.OnDefinition
	log          *zap.SugaredLogger
}

// shouldRun determines if a job should be triggered based on event, branch, and file changes.
func (j *Job) shouldRun(eventName string, targetBranch string, changedFiles []string) bool {
	if !j.matchBranchAndEvent(eventName, targetBranch) {
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

// matchesBranchAndEvent checks if the current event and branch match the rules in the 'on' block.
func (j *Job) matchBranchAndEvent(eventName string, targetBranch string) bool {
	if j.OnRules == nil {
		return true
	}

	eventRules, eventDefined := j.OnRules[eventName]
	if !eventDefined {
		return false
	}

	if len(eventRules.Branches) == 0 {
		return true
	}

	for _, allowedBranch := range eventRules.Branches {
		if allowedBranch == targetBranch {
			return true
		}
	}

	return false
}

// matchesFiles checks if any changed files match the job's file patterns.
func (j *Job) matchesFiles(changedFiles []string) bool {
	if len(j.PathPatterns) == 0 {
		return true
	}

	for _, pattern := range j.PathPatterns {
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
func NewJobMatcher(definitions configloader.JobDefinitions, log *zap.SugaredLogger) JobMatcher {
	var jobs []Job
	for name, def := range definitions {
		jobs = append(jobs, Job{
			Name:         name,
			PathPatterns: def.Paths,
			OnRules:      def.On,
			log:          log,
		})
	}

	return JobMatcher{jobs: jobs, log: log}
}

// MatchJobs is the primary method that applies all filters to a list of changed files.
func (p *JobMatcher) MatchJobs(eventName string, targetBranch string, changedFiles []string) JobFiltersResult {
	result := JobFiltersResult{
		TriggeredJobKeys: []string{},
		JobTriggers:      make(map[string]bool),
	}

	for _, job := range p.jobs {
		shouldRun := job.shouldRun(eventName, targetBranch, changedFiles)
		result.JobTriggers[job.Name] = shouldRun
		if shouldRun {
			result.TriggeredJobKeys = append(result.TriggeredJobKeys, job.Name)
		}
	}

	return result
}
