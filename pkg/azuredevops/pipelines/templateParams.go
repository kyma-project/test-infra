package pipelines

import (
	"strconv"
)

type OCIImageBuilderTemplateParams map[string]string

func (p OCIImageBuilderTemplateParams) SetRepoName(repoName string) {
	p["RepoName"] = repoName
}
func (p OCIImageBuilderTemplateParams) SetRepoOwner(repoOwner string) {
	p["RepoOwner"] = repoOwner
}
func (p OCIImageBuilderTemplateParams) SetJobType(jobType string) {
	p["JobType"] = jobType
}
func (p OCIImageBuilderTemplateParams) SetPullNumber(pullNumber string) {
	p["PullNumber"] = pullNumber
}
func (p OCIImageBuilderTemplateParams) SetPullBaseSHA(pullBaseSHA string) {
	p["PullBaseSHA"] = pullBaseSHA
}
func (p OCIImageBuilderTemplateParams) SetImageName(name string) {
	p["ImageName"] = name
}
func (p OCIImageBuilderTemplateParams) SetDockerfilePath(dockerfile string) {
	p["DockerfilePath"] = dockerfile
}
func (p OCIImageBuilderTemplateParams) SetBuildContext(context string) {
	p["BuildContext"] = context
}
func (p OCIImageBuilderTemplateParams) SetExportTags(exportTags bool) {
	p["ExportTags"] = strconv.FormatBool(exportTags)
}
func (p OCIImageBuilderTemplateParams) SetPlatforms(platforms string) {
	p["Platforms"] = platforms
}
func (p OCIImageBuilderTemplateParams) SetBuildArgs(buildArgs string) {
	p["BuildArgs"] = buildArgs
}
func (p OCIImageBuilderTemplateParams) SetImageTags(imageTags string) {
	p["ImageTags"] = imageTags
}
