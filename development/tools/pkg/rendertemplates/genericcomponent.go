package rendertemplates

import (
	"strings"
)

// copies values from the original jobConfig into the generated ones
func (j *Job) appendCommonValues(repo Repo, genericJob Job) {

	j.JobConfig["path_alias"] = repo.RepoName

	// copy all values except path
	for name, val := range genericJob.JobConfig {
		if name != "path" {
			j.JobConfig[name] = val
		}
	}
	j.InheritedConfigs.Local = genericJob.InheritedConfigs.Local
}

// GenerateComponentJobs generates jobs for components
func (r *RenderConfig) GenerateComponentJobs(global map[string]interface{}) {
	if present := len(r.JobConfigs); present > 0 {
		for repoIndex, repo := range r.JobConfigs {
			var jobs []Job
			hasComponentJobs := false

			for _, job := range repo.Jobs {
				// check if the jobConfig is in expected for component jobs format
				if job.JobConfig["name"] == nil && job.JobConfig["path"] != nil {
					hasComponentJobs = true
					// generate component jobs

					repository := strings.Split(repo.RepoName, "/")[2]
					nameSuffix := repository + "-" + strings.Replace(job.JobConfig["path"].(string), "/", "-", -1)

					// generate pre- and post-submit jobs for the next release
					var preSubmit Job
					preSubmit.JobConfig = make(map[string]interface{})
					preSubmit.appendCommonValues(repo, job)
					preSubmit.JobConfig["name"] = "pre-" + nameSuffix
					preSubmit.InheritedConfigs.Global = append(job.InheritedConfigs.Global, "jobConfig_presubmit", "extra_refs_test-infra")
					jobs = append(jobs, preSubmit)

					var postSubmit Job
					postSubmit.JobConfig = make(map[string]interface{})
					postSubmit.appendCommonValues(repo, job)
					postSubmit.JobConfig["name"] = "post-" + nameSuffix
					postSubmit.InheritedConfigs.Global = append(job.InheritedConfigs.Global, "jobConfig_postsubmit", "extra_refs_test-infra", "disable_testgrid")
					jobs = append(jobs, postSubmit)

					// check if we have to generate jobs for the previous supported releases
					if job.JobConfig["skipReleaseJobs"] == nil || job.JobConfig["skipReleaseJobs"].(string) != "true" {
						for _, currentRelease := range global["releases"].([]interface{}) {
							rel := currentRelease.(string)
							nameRelease := "rel" + strings.Replace(rel, ".", "", -1)
							commonRelBranches := []string{"release-" + rel}
							commonExtrarefsTestInfra := map[string]interface{}{"test-infra": []map[string]interface{}{{"org": "kyma-project", "repo": "test-infra", "path_alias": "github.com/kyma-project/test-infra", "base_ref": "release-" + rel}}}

							var preSubmitRel Job
							preSubmitRel.JobConfig = make(map[string]interface{})
							preSubmitRel.appendCommonValues(repo, job)
							preSubmitRel.JobConfig["name"] = "pre-" + nameRelease + "-" + nameSuffix
							preSubmitRel.JobConfig["branches"] = commonRelBranches
							preSubmitRel.JobConfig["extra_refs"] = commonExtrarefsTestInfra

							// let MergeConfig know to which release compare against
							preSubmitRel.JobConfig["release_current"] = rel
							preSubmitRel.InheritedConfigs.Global = append(job.InheritedConfigs.Global, "jobConfig_presubmit")
							jobs = append(jobs, preSubmitRel)

							var postSubmitRel Job
							postSubmitRel.JobConfig = make(map[string]interface{})
							postSubmitRel.appendCommonValues(repo, job)
							postSubmitRel.JobConfig["name"] = "post-" + nameRelease + "-" + nameSuffix
							postSubmitRel.JobConfig["branches"] = commonRelBranches
							postSubmitRel.JobConfig["extra_refs"] = commonExtrarefsTestInfra
							postSubmitRel.JobConfig["release_current"] = rel
							postSubmitRel.InheritedConfigs.Global = append(job.InheritedConfigs.Global, "jobConfig_postsubmit", "disable_testgrid")
							jobs = append(jobs, postSubmitRel)
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

	}
}
