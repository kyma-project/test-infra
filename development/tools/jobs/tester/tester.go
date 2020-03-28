package tester

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/kube"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

const (

	// ImageGolangBuildpackLatest means Golang buildpack image
	ImageGolangBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd"
	// ImageGolangBuildpack1_11 means Golang buildpack image with Go 1.11.*
	ImageGolangBuildpack1_11 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11"
	// ImageGolangBuildpack1_12 means Golang buildpack image with Go 1.12.*
	ImageGolangBuildpack1_12 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.12"
	// ImageGolangBuildpack1_13 means Golang buildpack image with Go 1.13.*
	ImageGolangBuildpack1_13 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.13"
	// ImageGolangKubebuilderBuildpackLatest means Golang buildpack with Kubebuilder image
	ImageGolangKubebuilderBuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder:v20190208-813daef"
	// ImageGolangKubebuilder2BuildpackLatest means Golang buildpack with Kubebuilder2 image
	ImageGolangKubebuilder2BuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20190823-24e14d1"
	// ImageNode10Buildpack means Node.js buildpack image (node v10)
	ImageNode10Buildpack = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node:v20181130-b28250b"
	// ImageNodeBuildpackLatest means Node.js buildpack image (node v12)
	ImageNodeBuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-node:v20191009-19b4b28"
	// ImageNodeChromiumBuildpackLatest means Node.js + Chromium buildpack image
	ImageNodeChromiumBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node-chromium:v20181207-d46c013"
	// ImageBootstrapLatest means Bootstrap image
	ImageBootstrapLatest = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181121-f3ea5ce"
	// ImageBootstrap20181204 represents boostrap image published on 2018.12.04
	ImageBootstrap20181204 = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181204-a6e79be"
	// ImageBootstrap20190604 represents boostrap image published on 2019.06.04
	ImageBootstrap20190604 = "eu.gcr.io/kyma-project/test-infra/bootstrap:v20190604-d08e7fe"
	// ImageBootstrap001 represents version 0.0.1 of bootstrap image
	ImageBootstrap001 = "eu.gcr.io/kyma-project/prow/bootstrap:0.0.1"
	// ImageKymaClusterInfraLatest represents boostrap image published on 20.11.2019
	ImageKymaClusterInfraK14    = "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20200124-8f253e51"
	ImageKymaClusterInfraK16    = "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20200206-22eb97a4"
	ImageKymaClusterInfraLatest = "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20200206-22eb97a4"
	// ImageKymaClusterInfra20190528 represents boostrap image published on 28.05.2019
	ImageKymaClusterInfra20190528 = "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190528-8897828"
	// ImageBootstrapHelm20181121 represents verion of bootstrap-helm image
	ImageBootstrapHelm20181121 = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap-helm:v20181121-f2f12bc"
	ImageBootstrapHelm20191227 = "eu.gcr.io/kyma-project/test-infra/bootstrap-helm:v20191227-cca719e8"
	// ImageGolangToolboxLatest represents the latest version of the golang buildpack toolbox
	ImageGolangToolboxLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20191004-f931536"

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
		if job.Name == name {
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
		if job.Name == name && job.RunsAgainstBranch(branch) {
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

// AssertThatHasExtraRefTestInfraWithSHA checks if job has configured extra ref to test-infra repository with appropriate sha
func AssertThatHasExtraRefTestInfraWithSHA(t *testing.T, in config.UtilityConfig, expectedBaseRef, expectedBaseSHA string) {
	for _, curr := range in.ExtraRefs {
		if curr.PathAlias == "github.com/kyma-project/test-infra" &&
			curr.Org == "kyma-project" &&
			curr.Repo == "test-infra" &&
			curr.BaseRef == expectedBaseRef &&
			curr.BaseSHA == expectedBaseSHA {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Job has not configured extra ref to test-infra repository with base ref set to [%s] sha", expectedBaseSHA))
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
func AssertThatHasPresets(t *testing.T, in config.JobBase, expected ...preset.Preset) {
	for _, p := range expected {
		require.Equal(t, "true", in.Labels[string(p)], "missing preset [%v]", p)
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
