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

// GenerateComponentJobs generates jobs for components
func (r *RenderConfig) GenerateComponentJobs(global map[string]interface{}) {
	if present := len(r.JobConfigs); present > 0 {
		for repoIndex, repo := range r.JobConfigs {
			var jobs []Job
			hasComponentJobs := false

			for _, job := range repo.Jobs {
				// check if the jobConfig is a component job
				if job.JobConfig["name"] == nil && job.JobConfig["path"] != nil {
					hasComponentJobs = true
					// generate component jobs

					repository := strings.Split(repo.RepoName, "/")[2]
					nameSuffix := repository + "-" + strings.Replace(job.JobConfig["path"].(string), "/", "-", -1)

					// generate pre- and post-submit jobs for the next release
					if ReleaseMatches(global["nextRelease"], job.JobConfig["release_since"], job.JobConfig["release_until"]) {
						if len(job.jobConfigPre) > 0 {
							preSubmit := Job{}
							preSubmit.JobConfig = deepCopyConfigSet(job.jobConfigPre)
							preSubmit.JobConfig["name"] = "pre-" + nameSuffix
							jobs = append(jobs, preSubmit)
						}

						if len(job.jobConfigPost) > 0 {
							postSubmit := Job{}
							postSubmit.JobConfig = deepCopyConfigSet(job.jobConfigPost)
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

							if len(job.jobConfigPre) > 0 {
								preSubmitRel := Job{}
								preSubmitRel.JobConfig = deepCopyConfigSet(job.jobConfigPre)
								preSubmitRel.JobConfig["name"] = "pre-" + nameRelease + "-" + nameSuffix
								preSubmitRel.JobConfig["branches"] = commonRelBranches
								preSubmitRel.changeExtraRefsBase("release-" + rel)
								jobs = append(jobs, preSubmitRel)
							}

							if len(job.jobConfigPost) > 0 {
								postSubmitRel := Job{}
								postSubmitRel.JobConfig = deepCopyConfigSet(job.jobConfigPost)
								postSubmitRel.JobConfig["name"] = "post-" + nameRelease + "-" + nameSuffix
								postSubmitRel.JobConfig["branches"] = commonRelBranches
								postSubmitRel.changeExtraRefsBase("release-" + rel)
								jobs = append(jobs, postSubmitRel)
							}
						}
					}
				} else {
					// append the job to the list, making it possible to mix component job definitions and regular ones
					jobs = append(jobs, job)
				}
			}

			// replace jobs if there were generated ones, don't change anything otherwise
			if hasComponentJobs {
				r.JobConfigs[repoIndex].Jobs = jobs
			}
		}
		r.Values["JobConfigs"] = r.JobConfigs
	}
}
