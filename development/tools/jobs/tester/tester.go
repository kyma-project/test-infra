package tester

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"k8s.io/test-infra/prow/config"
	"os"
	"testing"
)

type Preset string

const (
	PresetDindEnabled    Preset = "preset-dind-enabled"
	PresetGcrPush        Preset = "preset-sa-gcr-push"
	PresetDockerPushRepo Preset = "preset-docker-push-repository"
	PresetBuildPr        Preset = "preset-build-pr"
	PresetBuildMaster    Preset = "preset-build-master"
	PresetBuildRelease   Preset = "preset-build-release"

	ImageGolangBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd"
	EnvSourcesDir              = "SOURCES_DIR"
)

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

func AssertThatHasPresets(t *testing.T, in config.JobBase, expected ... Preset) {
	for _, p := range expected {
		assert.Equal(t, "true", in.Labels[string(p)], "missing preset [%s]", p)
	}
}

func AssertThatRunIfChanged(t *testing.T, p config.Presubmit, changedFile string) {
	sl := []config.Presubmit{p}
	require.NoError(t, config.SetPresubmitRegexes(sl))
	assert.True(t, sl[0].RunsAgainstChanges([]string{changedFile}), "missed change [%s]", changedFile)

}
