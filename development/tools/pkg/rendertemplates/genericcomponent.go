package rendertemplates

import (
	"strings"
)

// changeExtraRefsBase changes base_ref to base string for each extra_ref
func (j *Job) changeExtraRefsBase(base string) {
	if j.JobConfig["extra_refs"] != nil {
		for extraRefIndex := range j.JobConfig["extra_refs"].(map[interface{}]interface{}) {
			j.JobConfig["extra_refs"].(map[interface{}]interface{})[extraRefIndex].([]interface{})[0].(map[interface{}]interface{})["base_ref"] = base
		}
	}
}

// AppendJobs generates jobs for components if necessary and appends pre/post/main jobs to the list if they exist
func (r *RenderConfig) AppendJobs(global map[string]interface{}) {
	if present := len(r.JobConfigs); present > 0 {
		for repoIndex, repo := range r.JobConfigs {
			var jobs []Job

			for _, job := range repo.Jobs {
				// check if the jobConfig is a component job
				if job.JobConfig["name"] == nil && job.JobConfig["path"] != nil {
					// generate component jobs
					componentJobs := generateComponentJobs(global, repo, job)
					jobs = append(jobs, componentJobs...)
				} else {
					// append the job to the list, making it possible to mix component job definitions and regular ones in one data file
					if len(job.JobConfig) > 0 {
						jobs = append(jobs, job)
					}

					if len(job.JobConfigPre) > 0 {
						preSubmit := Job{}
						preSubmit.JobConfig = deepCopyConfigSet(job.JobConfigPre)
						jobs = append(jobs, preSubmit)
					}

					if len(job.JobConfigPost) > 0 {
						postSubmit := Job{}
						postSubmit.JobConfig = deepCopyConfigSet(job.JobConfigPost)
						jobs = append(jobs, postSubmit)
					}
				}
			}

			r.JobConfigs[repoIndex].Jobs = jobs
		}
		r.Values["JobConfigs"] = r.JobConfigs
	}
}

// generateComponentJobs generates jobs for components
func generateComponentJobs(global map[string]interface{}, repo Repo, job Job) []Job {
	var jobs []Job
	// generate component jobs

	// get last element of repoName, this way "github.com/" part can be omitted
	repoNameSlice := strings.Split(repo.RepoName, "/")
	repository := repoNameSlice[len(repoNameSlice)-1]
	nameSuffix := repository + "-" + strings.Replace(job.JobConfig["path"].(string), "/", "-", -1)

	// generate pre- and post-submit jobs for the next release
	if ReleaseMatches(global["nextRelease"], job.JobConfig["release_since"], job.JobConfig["release_until"]) {
		if len(job.JobConfigPre) > 0 {
			preSubmit := Job{}
			preSubmit.JobConfig = deepCopyConfigSet(job.JobConfigPre)
			preSubmit.JobConfig["name"] = "pre-" + nameSuffix
			jobs = append(jobs, preSubmit)
		}

		if len(job.JobConfigPost) > 0 {
			postSubmit := Job{}
			postSubmit.JobConfig = deepCopyConfigSet(job.JobConfigPost)
			postSubmit.JobConfig["name"] = "post-" + nameSuffix
			jobs = append(jobs, postSubmit)
		}
	}

	// check if we have to generate jobs for the previous supported releases
	if job.JobConfig["skipReleaseJobs"] == nil || job.JobConfig["skipReleaseJobs"].(string) != "true" {

		matchingReleases := MatchingReleases(global["releases"].([]interface{}), job.JobConfig["release_since"], job.JobConfig["release_until"])
		for _, currentRelease := range matchingReleases {
			rel := currentRelease.(string)
			nameRelease := "rel" + strings.Replace(rel, ".", "", -1)
			commonRelBranches := []string{"release-" + rel}

			if len(job.JobConfigPre) > 0 {
				preSubmitRel := Job{}
				preSubmitRel.JobConfig = deepCopyConfigSet(job.JobConfigPre)
				preSubmitRel.JobConfig["name"] = "pre-" + nameRelease + "-" + nameSuffix
				preSubmitRel.JobConfig["branches"] = commonRelBranches
				preSubmitRel.changeExtraRefsBase("release-" + rel)
				jobs = append(jobs, preSubmitRel)
			}

			if len(job.JobConfigPost) > 0 {
				postSubmitRel := Job{}
				postSubmitRel.JobConfig = deepCopyConfigSet(job.JobConfigPost)
				postSubmitRel.JobConfig["name"] = "post-" + nameRelease + "-" + nameSuffix
				postSubmitRel.JobConfig["branches"] = commonRelBranches
				postSubmitRel.changeExtraRefsBase("release-" + rel)
				jobs = append(jobs, postSubmitRel)
			}
		}
	}
	return jobs
}
