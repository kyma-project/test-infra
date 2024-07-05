// TODO: Add more structured logging with debug severity to track execution in case of troubleshooting

package pipelines

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
)

// ErrRequiredParamNotSet is returned when the required template parameter is not set
type ErrRequiredParamNotSet string

// Error returns the error message with the parameter name.
// Example of usage: ErrRequiredParamNotSet("RepoName")
// Returns: "required parameter not set: RepoName"
func (e ErrRequiredParamNotSet) Error() string {
	return "required parameter not set: " + string(e)
}

// OCIImageBuilderTemplateParams holds parameters accepted by oci-image-builder ADO pipeline.
// TODO: Rename, remove Template, as this is are parameters for pipeline execution.
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

// SetWorkflowDispatchJobType sets required parameter JobType to workflow_dispatch.
func (p OCIImageBuilderTemplateParams) SetWorkflowDispatchJobType() {
	p["JobType"] = "workflow_dispatch"
}

// SetPullNumber sets optional parameter PullNumber.
func (p OCIImageBuilderTemplateParams) SetPullNumber(number string) {
	p["PullNumber"] = number
}

// SetBaseSHA sets required parameter BaseSHA.
// For presubmit job, this is the pull request base commit SHA with source code for building image for tests.
// For postsubmit job, this is the branch commit SHA with source code used for building image.
func (p OCIImageBuilderTemplateParams) SetBaseSHA(sha string) {
	// TODO: Rename key to BaseSHA
	p["PullBaseSHA"] = sha
}

// SetBaseRef sets required parameter BaseRef.
func (p OCIImageBuilderTemplateParams) SetBaseRef(ref string) {
	p["BaseRef"] = ref
}

// SetPullSHA sets optional parameter PullSHA.
// This is the pull request head commit SHA with source code for building image for tests.
func (p OCIImageBuilderTemplateParams) SetPullSHA(sha string) {
	// TODO: Rename key to PullSHA
	p["PullPullSHA"] = sha
}

// SetImageName sets required parameter ImageName.
// This is the name of the image to be built.
func (p OCIImageBuilderTemplateParams) SetImageName(name string) {
	// TODO: Rename key to ImageName
	p["Name"] = name
}

// SetDockerfilePath sets required parameter DockerfilePath.
// This is a path relative to the context directory path.
func (p OCIImageBuilderTemplateParams) SetDockerfilePath(path string) {
	// TODO: Rename key to DockerfilePath
	p["Dockerfile"] = path
}

// SetEnvFilePath sets required parameter EnvFile.
// This is a path relative to the context directory path.
func (p OCIImageBuilderTemplateParams) SetEnvFilePath(path string) {
	p["EnvFile"] = path
}

// SetBuildContext sets required parameter BuildContext.
// This is the path to the build context directory.
func (p OCIImageBuilderTemplateParams) SetBuildContext(context string) {
	// TODO: Rename key to BuildContext
	p["Context"] = context
}

// SetExportTags sets optional parameter ExportTags.
// If true, ADO pipeline will export tags names and values as builda args to the image build process.
func (p OCIImageBuilderTemplateParams) SetExportTags(export bool) {
	p["ExportTags"] = strconv.FormatBool(export)
}

// SetBuildArgs sets optional parameter BuildArgs.
// This parameter is used to provide additional arguments for image build.
func (p OCIImageBuilderTemplateParams) SetBuildArgs(args string) {
	p["BuildArgs"] = args
}

// SetImageTags sets optional parameter Tags.
// This parameter is used to provide additional tags for the image.
// The value is base64 encoded to avoid issues with special characters.
func (p OCIImageBuilderTemplateParams) SetImageTags(tags string) {
	// TODO: Rename key to ImageTags
	encodedTags := base64.StdEncoding.EncodeToString([]byte(tags))
	p["Tags"] = encodedTags
}

// SetUseKanikoConfigFromPR sets optional parameter UseKanikoConfigFromPR.
// If true, ADO pipeline will use a Kaniko config from PR.
// This is used for testing purposes.
func (p OCIImageBuilderTemplateParams) SetUseKanikoConfigFromPR(useKanikoFromPR bool) {
	p["UseKanikoConfigFromPR"] = strconv.FormatBool(useKanikoFromPR)
}

// SetAuthorization sets Authorization parameter.
// This parameter is used to provide authorization token when running in github actions
func (p OCIImageBuilderTemplateParams) SetAuthorization(authorizationToken string) {
	p["Authorization"] = authorizationToken
}

// Validate validates if required OCIImageBuilderTemplateParams are set.
// Returns ErrRequiredParamNotSet error if any required parameter is not set.
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
	if !slices.Contains([]string{"presubmit", "postsubmit", "workflow_dispatch"}, jobType) {
		return fmt.Errorf("JobType must be either presubmit, postsubmit or workflow_dispatch, got: %s", jobType)
	}
	if _, ok = p["PullBaseSHA"]; !ok {
		return ErrRequiredParamNotSet("BaseSHA")
	}
	if _, ok = p["Name"]; !ok {
		return ErrRequiredParamNotSet("ImageName")
	}
	if _, ok = p["Dockerfile"]; !ok {
		return ErrRequiredParamNotSet("DockerfilePath")
	}
	if _, ok = p["Context"]; !ok {
		return ErrRequiredParamNotSet("BuildContext")
	}
	return nil
}
