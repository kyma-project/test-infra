package tester

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/test-infra/prow/kube"

	"fmt"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/test-infra/prow/config"
)

// Preset represents a existing presets
type Preset string

const (
	// PresetDindEnabled means docker-in-docker preset
	PresetDindEnabled Preset = "preset-dind-enabled"
	// PresetGcrPush means GCR push service account
	PresetGcrPush Preset = "preset-sa-gcr-push"
	// PresetDockerPushRepo means Docker repository
	PresetDockerPushRepo Preset = "preset-docker-push-repository"
	// PresetDockerPushRepoTestInfra means Docker repository test-infra images
	PresetDockerPushRepoTestInfra Preset = "preset-docker-push-repository-test-infra"
	// PresetDockerPushRepoIncubator means Decker repository incubator images
	PresetDockerPushRepoIncubator Preset = "preset-docker-push-repository-incubator"
	// PresetDockerPushGlobalRepo means Decker global repository for images
	PresetDockerPushGlobalRepo Preset = "preset-docker-push-global-repository"
	// PresetBuildPr means PR environment
	PresetBuildPr Preset = "preset-build-pr"
	// PresetBuildMaster means master environment
	PresetBuildMaster Preset = "preset-build-master"
	// PresetBuildConsoleMaster means console master environment
	PresetBuildConsoleMaster Preset = "preset-build-console-master"
	// PresetBuildRelease means release environment
	PresetBuildRelease Preset = "preset-build-release"
	// PresetBotGithubToken means github token
	PresetBotGithubToken Preset = "preset-bot-github-token"
	// PresetBotGithubSSH means github ssh
	PresetBotGithubSSH Preset = "preset-bot-github-ssh"
	// PresetBotGithubIdentity means github identity
	PresetBotGithubIdentity Preset = "preset-bot-github-identity"
	// PresetWebsiteBotGithubToken means github token
	PresetWebsiteBotGithubToken Preset = "preset-website-bot-github-token"
	// PresetKymaGuardBotGithubToken represents the Kyma Guard Bot token for GitHub
	PresetKymaGuardBotGithubToken Preset = "preset-kyma-guard-bot-github-token"
	// PresetWebsiteBotGithubSSH means github ssh
	PresetWebsiteBotGithubSSH Preset = "preset-website-bot-github-ssh"
	// PresetWebsiteBotGithubIdentity means github identity
	PresetWebsiteBotGithubIdentity Preset = "preset-website-bot-github-identity"
	// PresetWebsiteBotZenHubToken means zenhub token
	PresetWebsiteBotZenHubToken Preset = "preset-website-bot-zenhub-token"
	// PresetSaGKEKymaIntegration means access to service account capable of creating clusters and related resources
	PresetSaGKEKymaIntegration = "preset-sa-gke-kyma-integration"
	// PresetGCProjectEnv means project name is injected as env variable
	PresetGCProjectEnv = "preset-gc-project-env"
	// PresetKymaBackupRestoreBucket means the bucket used for backups and restore in Kyma
	PresetKymaBackupRestoreBucket = "preset-kyma-backup-restore-bucket"
	// PresetKymaBackupCredentials means the credentials for the service account
	PresetKymaBackupCredentials = "preset-kyma-backup-credentials"

	// ImageGolangBuildpackLatest means Golang buildpack image
	ImageGolangBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd"
	// ImageGolangBuildpack1_11 means Golang buildpack image with Go 1.11.*
	ImageGolangBuildpack1_11 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11"
	// ImageGolangBuildpack1_12 means Golang buildpack image with Go 1.12.*
	ImageGolangBuildpack1_12 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12"
	// ImageGolangKubebuilderBuildpackLatest means Golang buildpack with Kubebuilder image
	ImageGolangKubebuilderBuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder:v20190208-813daef"
	// ImageGolangKubebuilder2BuildpackLatest means Golang buildpack with Kubebuilder2 image
	ImageGolangKubebuilder2BuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20190823-24e14d1"
	// ImageNodeBuildpackLatest means Node.js buildpack image
	ImageNodeBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node:v20181130-b28250b"
	// ImageNodeChromiumBuildpackLatest means Node.js + Chromium buildpack image
	ImageNodeChromiumBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node-chromium:v20181207-d46c013"
	// ImageBootstrapLatest means Bootstrap image
	ImageBootstrapLatest = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181121-f3ea5ce"
	// ImageBootstrap20181204 represents boostrap image published on 2018.12.04
	ImageBootstrap20181204 = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181204-a6e79be"
	// ImageBootstrap001 represents version 0.0.1 of bootstrap image
	ImageBootstrap001 = "eu.gcr.io/kyma-project/prow/bootstrap:0.0.1"
	// ImageKymaClusterInfra20190528 represents boostrap image published on 28.05.2019
	ImageKymaClusterInfra20190528 = "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190528-8897828"
	// ImageBootstrapHelm20181121 represents verion of bootstrap-helm image
	ImageBootstrapHelm20181121 = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap-helm:v20181121-f2f12bc"

	// KymaProjectDir means kyma project dir
	KymaProjectDir = "/home/prow/go/src/github.com/kyma-project"
	// KymaIncubatorDir means kyma incubator dir
	KymaIncubatorDir = "/home/prow/go/src/github.com/kyma-incubator"

	// BuildScriptDir means build script directory
	BuildScriptDir = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
	// GovernanceScriptDir means governance script directory
	GovernanceScriptDir = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/governance.sh"
	// MetadataGovernanceScriptDir means governance script directory
	MetadataGovernanceScriptDir = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/metadata-governance.sh"
)

type jobRunner interface {
	RunsAgainstChanges([]string) bool
}

// ReadJobConfig reads job configuration from file
func ReadJobConfig(fileName string) (config.JobConfig, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return config.JobConfig{}, errors.Wrapf(err, "while opening file [%s]", fileName)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return config.JobConfig{}, errors.Wrapf(err, "while reading file [%s]", fileName)
	}
	jobConfig := config.JobConfig{}
	if err = yaml.Unmarshal(b, &jobConfig); err != nil {
		return config.JobConfig{}, errors.Wrapf(err, "while unmarshalling file [%s]", fileName)
	}

	for _, v := range jobConfig.Presubmits {
		if err := config.SetPresubmitRegexes(v); err != nil {
			return config.JobConfig{}, errors.Wrap(err, "while setting presubmit regexes")
		}
	}

	for _, v := range jobConfig.Postsubmits {
		if err := config.SetPostsubmitRegexes(v); err != nil {
			return config.JobConfig{}, errors.Wrap(err, "while setting postsubmit regexes")
		}
	}
	return jobConfig, nil
}

// FindPresubmitJobByNameAndBranch finds presubmit job by name from provided jobs list
func FindPresubmitJobByNameAndBranch(jobs []config.Presubmit, name, branch string) *config.Presubmit {
	for _, job := range jobs {
		if job.Name == name && job.RunsAgainstBranch(branch) {
			return &job
		}
	}

	return nil
}

// FindPresubmitJobByName finds presubmit job by name from provided jobs list
func FindPresubmitJobByName(jobs []config.Presubmit, name string) *config.Presubmit {
	for _, job := range jobs {
		if job.Name == name  {
			return &job
		}
	}

	return nil
}

// GetReleaseJobName returns name of release job based on branch name by adding release prefix
func GetReleaseJobName(moduleName string, release *releases.SupportedRelease) string {
	return fmt.Sprintf("pre-%s-%s", release.JobPrefix(), moduleName)
}

// GetReleasePostSubmitJobName returns name of postsubmit job based on branch name
func GetReleasePostSubmitJobName(moduleName string, release *releases.SupportedRelease) string {
	return fmt.Sprintf("post-%s-%s", release.JobPrefix(), moduleName)
}

// FindPostsubmitJobByNameAndBranch finds postsubmit job by name from provided jobs list
func FindPostsubmitJobByNameAndBranch(jobs []config.Postsubmit, name, branch string) *config.Postsubmit {
	for _, job := range jobs {
		if job.Name == name && job.RunsAgainstBranch(branch){
			return &job
		}
	}

	return nil
}

// FindPostsubmitJobByNameAndBranch finds postsubmit job by name from provided jobs list
func FindPostsubmitJobByName(jobs []config.Postsubmit, name string) *config.Postsubmit {
	for _, job := range jobs {
		if job.Name == name {
			return &job
		}
	}

	return nil
}

// FindPeriodicJobByName finds periodic job by name from provided jobs list
func FindPeriodicJobByName(jobs []config.Periodic, name string) *config.Periodic {
	for _, job := range jobs {
		if job.Name == name {
			return &job
		}
	}

	return nil
}

// AssertThatHasExtraRefTestInfra checks if job has configured extra ref to test-infra repository
func AssertThatHasExtraRefTestInfra(t *testing.T, in config.UtilityConfig, expectedBaseRef string) {
	for _, curr := range in.ExtraRefs {
		if curr.PathAlias == "github.com/kyma-project/test-infra" &&
			curr.Org == "kyma-project" &&
			curr.Repo == "test-infra" &&
			curr.BaseRef == expectedBaseRef {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Job has not configured extra ref to test-infra repository with base ref set to [%s]", expectedBaseRef))
}

// AssertThatHasExtraRefs checks if UtilityConfig has repositories passed in argument defined
func AssertThatHasExtraRefs(t *testing.T, in config.UtilityConfig, repositories []string) {
	for _, repository := range repositories {
		for _, curr := range in.ExtraRefs {
			if curr.PathAlias == fmt.Sprintf("github.com/kyma-project/%s", repository) &&
				curr.Org == "kyma-project" &&
				curr.Repo == repository &&
				curr.BaseRef == "master" {
				return
			}
		}
		assert.FailNow(t, fmt.Sprintf("Job has not configured %s as a extra ref", repository))
	}
}

// AssertThatHasPresets checks if JobBase has expected labels
func AssertThatHasPresets(t *testing.T, in config.JobBase, expected ...Preset) {
	for _, p := range expected {
		assert.Equal(t, "true", in.Labels[string(p)], "missing preset [%s]", p)
	}
}

// AssertThatJobRunIfChanged checks if job that has specified run_if_changed parameter will be triggered by changes in specified file.
func AssertThatJobRunIfChanged(t *testing.T, p jobRunner, changedFile string) {
	assert.True(t, p.RunsAgainstChanges([]string{changedFile}), "missed change [%s]", changedFile)
}

// AssertThatJobDoesNotRunIfChanged checks if job that has specified run_if_changed parameter will not be triggered by changes in specified file.
func AssertThatJobDoesNotRunIfChanged(t *testing.T, p jobRunner, changedFile string) {
	assert.False(t, p.RunsAgainstChanges([]string{changedFile}), "triggered by changed file [%s]", changedFile)
}

// AssertThatHasCommand checks if job has
func AssertThatHasCommand(t *testing.T, command []string) {
	assert.Equal(t, []string{BuildScriptDir}, command)
}

// AssertThatExecGolangBuildpack checks if job executes golang buildpack
func AssertThatExecGolangBuildpack(t *testing.T, job config.JobBase, img string, args ...string) {
	assert.Len(t, job.Spec.Containers, 1)
	assert.Equal(t, img, job.Spec.Containers[0].Image)
	assert.Len(t, job.Spec.Containers[0].Command, 1)
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh", job.Spec.Containers[0].Command[0])
	assert.Equal(t, args, job.Spec.Containers[0].Args)
	assert.True(t, *job.Spec.Containers[0].SecurityContext.Privileged)
}

// AssertThatSpecifiesResourceRequests checks if resources requests for memory and cpu are specified
func AssertThatSpecifiesResourceRequests(t *testing.T, job config.JobBase) {
	assert.Len(t, job.Spec.Containers, 1)
	assert.False(t, job.Spec.Containers[0].Resources.Requests.Memory().IsZero())
	assert.False(t, job.Spec.Containers[0].Resources.Requests.Cpu().IsZero())
}

// AssertThatContainerHasEnv checks if container has specified given environment variable
func AssertThatContainerHasEnv(t *testing.T, cont kube.Container, expName, expValue string) {
	for _, env := range cont.Env {
		if env.Name == expName && env.Value == expValue {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Container [%s] does not have environment variable [%s] with value [%s]", cont.Name, expName, expValue))
}

// AssertThatContainerHasEnvFromSecret checks if container has specified given environment variable
func AssertThatContainerHasEnvFromSecret(t *testing.T, cont kube.Container, expName, expSecretName, expSecretKey string) {
	for _, env := range cont.Env {
		if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == expSecretName && env.ValueFrom.SecretKeyRef.Key == expSecretKey {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Container [%s] does not have environment variable [%s] with value from secret [name: %s, key: %s]", cont.Name, expName, expSecretName, expSecretKey))
}
