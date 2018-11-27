package tester

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// PresetBuildPr means PR environment
	PresetBuildPr Preset = "preset-build-pr"
	// PresetBuildMaster means master environment
	PresetBuildMaster Preset = "preset-build-master"
	// PresetBuildRelease means release environment
	PresetBuildRelease Preset = "preset-build-release"

	// ImageGolangBuildpackLatest means Golang buildpack image
	ImageGolangBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd"
	// ImageNodeBuildpackLatest means Node.js buildpack image
	ImageNodeBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node:v20181119-afd3fbd"
	// ImageBootstrapLatest means Bootstrap image
	ImageBootstrapLatest = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181121-f3ea5ce"

	// BuildScriptDir means build script directory
	BuildScriptDir = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
)

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
	return jobConfig, nil
}

// FindPresubmitJobByName finds presubmit job by name from provided jobs list
func FindPresubmitJobByName(jobs []config.Presubmit, name string) *config.Presubmit {
	for _, job := range jobs {
		if job.Name == name {
			return &job
		}
	}

	return nil
}

// FindPostsubmitJobByName finds postsubmit job by name from provided jobs list
func FindPostsubmitJobByName(jobs []config.Postsubmit, name string) *config.Postsubmit {
	for _, job := range jobs {
		if job.Name == name {
			return &job
		}
	}

	return nil
}

// AssertThatHasExtraRefTestInfra checks if UtilityConfig has test-infra repository defined
func AssertThatHasExtraRefTestInfra(t *testing.T, in config.UtilityConfig) {
	for _, curr := range in.ExtraRefs {
		if curr.PathAlias == "github.com/kyma-project/test-infra" &&
			curr.Org == "kyma-project" &&
			curr.Repo == "test-infra" &&
			curr.BaseRef == "master" {
			return
		}
	}
	assert.FailNow(t, "Job has not configured test-infra as a extra ref")
}

// AssertThatHasPresets checks if JobBase has expected labels
func AssertThatHasPresets(t *testing.T, in config.JobBase, expected ...Preset) {
	for _, p := range expected {
		assert.Equal(t, "true", in.Labels[string(p)], "missing preset [%s]", p)
	}
}

// AssertThatJobRunIfChanged checks if Presubmit has run_if_changed parameter
func AssertThatJobRunIfChanged(t *testing.T, p config.Presubmit, changedFile string) {
	sl := []config.Presubmit{p}
	require.NoError(t, config.SetPresubmitRegexes(sl))
	assert.True(t, sl[0].RunsAgainstChanges([]string{changedFile}), "missed change [%s]", changedFile)

}

// AssertThatHasCommand checks if job has
func AssertThatHasCommand(t *testing.T, command []string) {
	assert.Equal(t, []string{BuildScriptDir}, command)
}
