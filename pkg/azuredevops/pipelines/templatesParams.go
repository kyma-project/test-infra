// TODO: Add more structured logging with debug severity to track execution in case of troubleshooting

package pipelines

import (
	"fmt"
	"strconv"
)

// ErrRequiredParamNotSet is returned when required parameter is not set
type ErrRequiredParamNotSet string

func (e ErrRequiredParamNotSet) Error() string {
	return "required parameter not set: " + string(e)
}

// OCIImageBuilderTemplateParams is a map of parameters for OCIImageBuilderTemplate
type OCIImageBuilderTemplateParams map[string]string

// SetRepoName sets required parameter RepoName
func (p OCIImageBuilderTemplateParams) SetRepoName(repo string) {
	p["RepoName"] = repo
}

// SetRepoOwner sets required parameter RepoOwner
func (p OCIImageBuilderTemplateParams) SetRepoOwner(owner string) {
	p["RepoOwner"] = owner
}

// SetPresubmitJobType sets required parameter JobType to presubmit.
func (p OCIImageBuilderTemplateParams) SetPresubmitJobType() {
	p["JobType"] = "presubmit"
}

// SetPostsubmitJobType sets required parameter JobType to postsubmit.
func (p OCIImageBuilderTemplateParams) SetPostsubmitJobType() {
	p["JobType"] = "postsubmit"
}

// SetPullNumber sets optional parameter PullNumber.
func (p OCIImageBuilderTemplateParams) SetPullNumber(number string) {
	p["PullNumber"] = number
}

// SetBaseSHA sets required parameter BaseSHA.
func (p OCIImageBuilderTemplateParams) SetBaseSHA(sha string) {
	// TODO: Rename key to BaseSHA
	p["PullBaseSHA"] = sha
}

// SetPullSHA sets optional parameter PullSHA.
func (p OCIImageBuilderTemplateParams) SetPullSHA(sha string) {
	// TODO: Rename key to PullSHA
	p["PullPullSHA"] = sha
}

// SetImageName sets required parameter ImageName.
func (p OCIImageBuilderTemplateParams) SetImageName(name string) {
	// TODO: Rename key to ImageName
	p["Name"] = name
}

// SetDockerfilePath sets required parameter DockerfilePath.
func (p OCIImageBuilderTemplateParams) SetDockerfilePath(path string) {
	// TODO: Rename key to DockerfilePath
	p["Dockerfile"] = path
}

// SetBuildContext sets required parameter BuildContext.
func (p OCIImageBuilderTemplateParams) SetBuildContext(context string) {
	// TODO: Rename key to BuildContext
	p["Context"] = context
}

// SetExportTags sets optional parameter ExportTags.
func (p OCIImageBuilderTemplateParams) SetExportTags(export bool) {
	p["ExportTags"] = strconv.FormatBool(export)
}

// SetBuildArgs sets optional parameter BuildArgs.
func (p OCIImageBuilderTemplateParams) SetBuildArgs(args string) {
	p["BuildArgs"] = args
}

// SetImageTags sets optional parameter ImageTags.
func (p OCIImageBuilderTemplateParams) SetImageTags(tags string) {
	// TODO: Rename key to ImageTags
	p["Tags"] = tags
}

// Validate validates if required OCIImageBuilderTemplateParams are set
func (p OCIImageBuilderTemplateParams) Validate() error {
	var (
		jobType string
		ok      bool
	)
	if _, ok = p["RepoName"]; !ok {
		return ErrRequiredParamNotSet("RepoName")
	}
	if _, ok = p["RepoOwner"]; !ok {
		return ErrRequiredParamNotSet("RepoOwner")
	}
	if jobType, ok = p["JobType"]; !ok {
		return ErrRequiredParamNotSet("JobType")
	}
	if jobType != "presubmit" && jobType != "postsubmit" {
		return fmt.Errorf("JobType must be either presubmit or postsubmit, got: %s", jobType)
	}
	if _, ok = p["PullBaseSHA"]; !ok {
		return ErrRequiredParamNotSet("BaseSHA")
	}
	if _, ok = p["ImageName"]; !ok {
		return ErrRequiredParamNotSet("ImageName")
	}
	if _, ok = p["DockerfilePath"]; !ok {
		return ErrRequiredParamNotSet("DockerfilePath")
	}
	if _, ok = p["BuildContext"]; !ok {
		return ErrRequiredParamNotSet("BuildContext")
	}
	return nil
}
