package rendertemplates

import (
	"log"
	"strings"
)

func (j *Job) appendCommonValues(r *RenderConfig) {
	commonGlobalConfigs := []string{"jobConfig_default", "image_buildpack-golang-kubebuilder2"}

	if r.Values["pushRepository"] == nil {
		log.Fatalln("Component jobs: missing \"pushrepository\" value.")
	}
	commonLabels := map[string]interface{}{"preset-dind-enabled": "true", "preset-sa-gcr-push": "true", "preset-docker-push-repository-" + r.Values["pushRepository"].(string): "true"}

	commonCommand := "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"
	commonArgs := []string{"/home/prow/go/src/" + r.Values["repository"].(string) + "/" + r.Values["path"].(string)}
	var commonOptional bool
	if r.Values["optional"] != nil {
		commonOptional = r.Values["optional"].(bool)
	}

	if r.Values["additionalRunIfChanged"] == nil {
		log.Fatalln("Component jobs: missing \"additionalRunIfChanged\" value.")
	}
	additionalRunIfChanged := make([]string, len(r.Values["additionalRunIfChanged"].([]interface{})))
	for i, add := range r.Values["additionalRunIfChanged"].([]interface{}) {
		additionalRunIfChanged[i] = add.(string)
	}
	commonRunIfChanged := "^" + r.Values["path"].(string) + "/|" + strings.Join(additionalRunIfChanged, "|")

	j.JobConfig["command"] = commonCommand
	j.JobConfig["args"] = commonArgs
	j.JobConfig["labels"] = commonLabels
	j.JobConfig["path_alias"] = r.Values["repository"]
	j.JobConfig["run_if_changed"] = commonRunIfChanged
	j.JobConfig["optional"] = commonOptional
	j.InheritedConfigs.Global = commonGlobalConfigs
}

func (r *RenderConfig) GenerateComponentJobs(global map[string]interface{}, globalConfigSets map[string]ConfigSet) {
	jobs := Repo{RepoName: r.Values["repository"].(string)}
	repository := strings.Split(r.Values["repository"].(string), "/")[2]
	nameSuffix := repository + "-" + strings.Replace(r.Values["path"].(string), "/", "-", -1)

	var preSubmit Job
	preSubmit.JobConfig = make(map[string]interface{})
	preSubmit.appendCommonValues(r)
	preSubmit.JobConfig["name"] = "pre-" + nameSuffix
	preSubmit.InheritedConfigs.Global = append(preSubmit.InheritedConfigs.Global, "jobConfig_presubmit", "extra_refs_test-infra")
	jobs.Jobs = append(jobs.Jobs, preSubmit)

	var postSubmit Job
	postSubmit.JobConfig = make(map[string]interface{})
	postSubmit.appendCommonValues(r)
	postSubmit.JobConfig["name"] = "post-" + nameSuffix
	postSubmit.InheritedConfigs.Global = append(postSubmit.InheritedConfigs.Global, "jobConfig_postsubmit", "extra_refs_test-infra", "disable_testgrid")
	jobs.Jobs = append(jobs.Jobs, postSubmit)

	// generate jobs for the previous releases
	if r.Values["skipReleaseJobs"] == nil || r.Values["skipReleaseJobs"].(string) != "true" {
		for _, currentRelease := range global["releases"].([]interface{}) {
			rel := currentRelease.(string)
			nameRelease := "rel" + strings.Replace(rel, ".", "", -1)
			commonRelBranches := []string{"release-" + rel}
			commonExtrarefsTestInfra := map[string]interface{}{"test-infra": []map[string]interface{}{{"org": "kyma-project", "repo": "test-infra", "path_alias": "github.com/kyma-project/test-infra", "base_ref": "release-" + rel}}}

			var preSubmitRel Job
			preSubmitRel.JobConfig = make(map[string]interface{})
			preSubmitRel.appendCommonValues(r)
			preSubmitRel.JobConfig["name"] = "pre-" + nameRelease + "-" + nameSuffix
			preSubmitRel.JobConfig["branches"] = commonRelBranches
			preSubmitRel.JobConfig["extra_refs"] = commonExtrarefsTestInfra
			preSubmitRel.InheritedConfigs.Global = append(preSubmitRel.InheritedConfigs.Global, "jobConfig_presubmit")
			jobs.Jobs = append(jobs.Jobs, preSubmitRel)

			var postSubmitRel Job
			postSubmitRel.JobConfig = make(map[string]interface{})
			postSubmitRel.appendCommonValues(r)
			postSubmitRel.JobConfig["name"] = "post-" + nameRelease + "-" + nameSuffix
			postSubmitRel.JobConfig["branches"] = commonRelBranches
			postSubmitRel.JobConfig["extra_refs"] = commonExtrarefsTestInfra
			postSubmitRel.InheritedConfigs.Global = append(postSubmitRel.InheritedConfigs.Global, "jobConfig_postsubmit", "disable_testgrid")
			jobs.Jobs = append(jobs.Jobs, postSubmitRel)
		}
	}

	r.JobConfigs = append(r.JobConfigs, jobs)
}
