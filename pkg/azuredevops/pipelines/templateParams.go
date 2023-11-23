package pipelines

import (
	"strconv"
)

type OCIImageBuilderTemplateParams map[string]string

func (p OCIImageBuilderTemplateParams) SetRepoName(repo string) {
	p["RepoName"] = repo
}
func (p OCIImageBuilderTemplateParams) SetRepoOwner(owner string) {
	p["RepoOwner"] = owner
}
func (p OCIImageBuilderTemplateParams) SetJobType(jobType string) {
	p["JobType"] = jobType
}
func (p OCIImageBuilderTemplateParams) SetPullNumber(number string) {
	p["PullNumber"] = number
}
func (p OCIImageBuilderTemplateParams) SetPullBaseSHA(sha string) {
	p["PullBaseSHA"] = sha
}
func (p OCIImageBuilderTemplateParams) SetImageName(name string) {
	p["ImageName"] = name
}
func (p OCIImageBuilderTemplateParams) SetDockerfilePath(path string) {
	p["DockerfilePath"] = path
}
func (p OCIImageBuilderTemplateParams) SetBuildContext(context string) {
	p["BuildContext"] = context
}
func (p OCIImageBuilderTemplateParams) SetExportTags(export bool) {
	p["ExportTags"] = strconv.FormatBool(export)
}
func (p OCIImageBuilderTemplateParams) SetBuildArgs(args string) {
	p["BuildArgs"] = args
}
func (p OCIImageBuilderTemplateParams) SetImageTags(tags string) {
	p["ImageTags"] = tags
}
