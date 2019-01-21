package tester

import (
	"io/ioutil"
	"os"
	"testing"

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
	// PresetSaGKEKymaIntegration means access to service account capable of creating clusters and related resources
	PresetSaGKEKymaIntegration = "preset-sa-gke-kyma-integration"
	// PresetGCProjectEnv means project name is injected as env variable
	PresetGCProjectEnv = "preset-gc-project-env"

	// ImageGolangBuildpackLatest means Golang buildpack image
	ImageGolangBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd"
	// ImageGolangBuildpack1_11 means Golang buildpack image with Go 1.11.*
	ImageGolangBuildpack1_11 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11"
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
	// ImageBootstrapHelm20181121 represents verion of bootstrap-helm image
	ImageBootstrapHelm20181121 = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap-helm:v20181121-f2f12bc"

	// KymaProjectDir means kyma project dir
	KymaProjectDir = "/home/prow/go/src/github.com/kyma-project"

	// BuildScriptDir means build script directory
	BuildScriptDir = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
	// GovernanceScriptDir means governance script directory
	GovernanceScriptDir = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/governance.sh"
)

type jobRunner interface {
	RunsAgainstChanges([]string) bool
}

// GetAllKymaReleaseBranches returns all supported kyma release branches
func GetAllKymaReleaseBranches() []string {
	return []string{"release-0.6"}
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

// FindPresubmitJobByName finds presubmit job by name from provided jobs list
func FindPresubmitJobByName(jobs []config.Presubmit, name, branch string) *config.Presubmit {
	for _, job := range jobs {
		if job.Name == name && job.RunsAgainstBranch(branch) {
			return &job
		}
	}

	return nil
}

// FindPostsubmitJobByName finds postsubmit job by name from provided jobs list
func FindPostsubmitJobByName(jobs []config.Postsubmit, name, branch string) *config.Postsubmit {
	for _, job := range jobs {
		if job.Name == name && job.RunsAgainstBranch(branch) {
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

// AssertThatHasCommand checks if job has
func AssertThatHasCommand(t *testing.T, command []string) {
	assert.Equal(t, []string{BuildScriptDir}, command)
}

// AssertThatExecGolangBuidlpack checks if job executes golang buildpack
func AssertThatExecGolangBuidlpack(t *testing.T, job config.JobBase, img string, args ...string) {
	assert.Len(t, job.Spec.Containers, 1)
	assert.Equal(t, img, job.Spec.Containers[0].Image)
	assert.Len(t, job.Spec.Containers[0].Command, 1)
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh", job.Spec.Containers[0].Command[0])
	assert.Equal(t, args, job.Spec.Containers[0].Args)
}

// AssertThatSpecifiesResourceRequests checks if resources requests for memory and cpu are specified
func AssertThatSpecifiesResourceRequests(t *testing.T, job config.JobBase) {
	assert.Len(t, job.Spec.Containers, 1)
	assert.False(t, job.Spec.Containers[0].Resources.Requests.Memory().IsZero())
	assert.False(t, job.Spec.Containers[0].Resources.Requests.Cpu().IsZero())

}
