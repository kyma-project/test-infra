package tester

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/config"
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
	// ImageGolangBuildpack1_14 means Golang buildpack image with Go 1.14.*
	ImageGolangBuildpack1_14 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.14"
	// ImageGolangKubebuilder2BuildpackLatest means Golang buildpack with Kubebuilder2 image
	ImageGolangKubebuilder2BuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20210407-70afe459"
	// ImageGolangKubebuilder2_20210216 contains the v20210216-1.14.0-rc1-535-gb45951c9 image
	ImageGolangKubebuilder2_20210216 = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20210216-1.14.0-rc1-535-gb45951c9"
	// ImageNode10Buildpack means Node.js buildpack image (node v10)
	ImageNode10Buildpack = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node:v20181130-b28250b"
	// ImageNodeBuildpackLatest means Node.js buildpack image (node v12)
	ImageNodeBuildpackLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-node:PR-2951"
	// ImageNodeChromiumBuildpackLatest means Node.js + Chromium buildpack image
	ImageNodeChromiumBuildpackLatest = "eu.gcr.io/kyma-project/prow/test-infra/buildpack-node-chromium:v20181207-d46c013"
	// ImageBootstrapLatest means Bootstrap image
	ImageBootstrapLatest = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181121-f3ea5ce"
	// ImageBootstrapTestInfraLatest means it's used in test-infra prowjob defs.
	ImageBootstrapTestInfraLatest = "eu.gcr.io/kyma-project/test-infra/bootstrap:PR-2951"
	// ImageBootstrapTestInfraCurrent means it's used in test-infra prowjob defs.
	ImageBootstrapTestInfraCurrent = "eu.gcr.io/kyma-project/test-infra/bootstrap:current"
	// ImageBootstrap20181204 represents boostrap image published on 2018.12.04
	ImageBootstrap20181204 = "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181204-a6e79be"
	// ImageBootstrap20190604 represents boostrap image published on 2019.06.04
	ImageBootstrap20190604 = "eu.gcr.io/kyma-project/test-infra/bootstrap:v20190604-d08e7fe"
	// ImageBootstrap001 represents version 0.0.1 of bootstrap image
	ImageBootstrap001 = "eu.gcr.io/kyma-project/prow/bootstrap:0.0.1"
	// ImageKymaIntegrationK14 represents kyma integration image with kubectl 1.14
	ImageKymaIntegrationK14 = "eu.gcr.io/kyma-project/test-infra/kyma-integration:v20200817-1.14.0-rc1-75-g625c1054-k8s1.14"
	// ImageKymaIntegrationK15 represents kyma integration image with kubectl 1.15
	ImageKymaIntegrationK15 = "eu.gcr.io/kyma-project/test-infra/kyma-integration:v20200817-1.14.0-rc1-75-g625c1054-k8s1.15"
	// ImageKymaIntegrationLatest represents kyma integration image with kubectl 1.16
	ImageKymaIntegrationLatest = "eu.gcr.io/kyma-project/test-infra/kyma-integration:v20210519-51c0242d-k8s1.16"
	// ImageGolangToolboxLatest represents the latest version of the golang buildpack toolbox
	ImageGolangToolboxLatest = "eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20210323-bfa93550"
	// ImageProwToolsLatest represents the latest version of the prow-tools image
	ImageProwToolsLatest = "eu.gcr.io/kyma-project/test-infra/prow-tools:v20210401-294e46e5"
	// ImageProwToolsCurrent represents the latest version of the prow-tools image
	ImageProwToolsCurrent = "eu.gcr.io/kyma-project/test-infra/prow-tools:current"
	// KymaProjectDir means kyma project dir
	KymaProjectDir = "/home/prow/go/src/github.com/kyma-project"
	// KymaIncubatorDir means kyma incubator dir
	KymaIncubatorDir = "/home/prow/go/src/github.com/kyma-incubator"

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

	for _, v := range jobConfig.PresubmitsStatic {
		if err := config.SetPresubmitRegexes(v); err != nil {
			return config.JobConfig{}, errors.Wrap(err, "while setting presubmit regexes")
		}
	}

	for _, v := range jobConfig.PostsubmitsStatic {
		if err := config.SetPostsubmitRegexes(v); err != nil {
			return config.JobConfig{}, errors.Wrap(err, "while setting postsubmit regexes")
		}
	}
	return jobConfig, nil
}

// FindPresubmitJobByNameAndBranch finds presubmit job by name from provided jobs list
func FindPresubmitJobByNameAndBranch(jobs []config.Presubmit, name, branch string) *config.Presubmit {
	for _, job := range jobs {
		if job.Name == name && job.CouldRun(branch) {
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

/*
IfPresubmitShouldRunAgainstChanges determines if the given presubmit job should run against given list of files
by checking them against regular expression if present. If the state of the job execution could not be determined
the function returns default state def.
*/
func IfPresubmitShouldRunAgainstChanges(job config.Presubmit, def bool, changedFiles ...string) bool {
	if job.AlwaysRun {
		return true
	}

	changed := func() ([]string, error) {
		return changedFiles, nil
	}
	det, shouldRun, err := job.RegexpChangeMatcher.ShouldRun(changed)
	if err != nil {
		fmt.Printf("An error occured during IfPresubmitShouldRunAgainstChanges execution: %v", err)
		return false
	}
	if det {
		return shouldRun
	}
	return def
}

/*
IfPostsubmitShouldRunAgainstChanges determines if the given postsubmit job should run against given list of files
by checking them against regular expression if present.
*/
func IfPostsubmitShouldRunAgainstChanges(job config.Postsubmit, changedFiles ...string) bool {
	changed := func() ([]string, error) {
		return changedFiles, nil
	}
	det, shouldRun, err := job.RegexpChangeMatcher.ShouldRun(changed)
	if err != nil {
		fmt.Printf("An error occured during IfPostsubmitShouldRunAgainstChanges execution: %v", err)
		return false
	}
	if det {
		return shouldRun
	}
	// postsubmits should run by default
	return true

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
		if job.Name == name && job.CouldRun(branch) {
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

// AssertThatHasExtraRef checks if UtilityConfig has ExtraRefs passed in argument defined
func AssertThatHasExtraRef(t *testing.T, in config.UtilityConfig, extraRefs []prowapi.Refs) {
	t.Helper()
	for _, ref := range extraRefs {
		assert.Contains(t, in.ExtraRefs, ref, fmt.Sprintf("\"%s\" ExtraRef not found in job", ref.Repo))
	}
}

// AssertThatHasExtraRepoRef checks if UtilityConfig has repositories passed in argument defined
func AssertThatHasExtraRepoRef(t *testing.T, in config.UtilityConfig, repositories []string) {
	t.Helper()
	var extraRefs []prowapi.Refs
	for _, repository := range repositories {
		extraRefs = append(extraRefs, prowapi.Refs{
			Org:       "kyma-project",
			Repo:      repository,
			BaseRef:   "master",
			PathAlias: fmt.Sprintf("github.com/kyma-project/%s", repository),
		})
	}
	AssertThatHasExtraRef(t, in, extraRefs)
}

// AssertThatHasExtraRepoRefCustom checks if UtilityConfig has repositories passed in argument defined with custom branches set
func AssertThatHasExtraRepoRefCustom(t *testing.T, in config.UtilityConfig, repositories []string, branches []string) {
	t.Helper()
	var extraRefs []prowapi.Refs
	for index, repository := range repositories {
		extraRefs = append(extraRefs, prowapi.Refs{
			Org:       "kyma-project",
			Repo:      repository,
			BaseRef:   branches[index],
			PathAlias: fmt.Sprintf("github.com/kyma-project/%s", repository),
		})
	}
	AssertThatHasExtraRef(t, in, extraRefs)
}

// AssertThatHasPresets checks if JobBase has expected labels
func AssertThatHasPresets(t *testing.T, in config.JobBase, expected ...preset.Preset) {
	for _, p := range expected {
		require.Equal(t, "true", in.Labels[string(p)], "missing preset [%v]", p)
	}
}

/*
AssertThatJobRunIfChanged checks if job that has specified run_if_changed parameter will be triggered by changes in specified file.
Deprecated: Please use IfPresubmitShouldRunAgainstChanges or IfPostsubmitShouldRunAgainstChanges for determining if job should run against given files.
*/
func AssertThatJobRunIfChanged(t *testing.T, p jobRunner, changedFile string) {
	assert.True(t, p.RunsAgainstChanges([]string{changedFile}), "missed change [%s]", changedFile)
}

/*
AssertThatJobDoesNotRunIfChanged checks if job that has specified run_if_changed parameter will not be triggered by changes in specified file.
Deprecated: Please use IfPresubmitShouldRunAgainstChanges or IfPostsubmitShouldRunAgainstChanges for determining if job should run against given files.
*/
func AssertThatJobDoesNotRunIfChanged(t *testing.T, p jobRunner, changedFile string) {
	assert.False(t, p.RunsAgainstChanges([]string{changedFile}), "triggered by changed file [%s]", changedFile)
}

// AssertThatExecGolangBuildpack checks if job executes golang buildpack
func AssertThatExecGolangBuildpack(t *testing.T, job config.JobBase, img string, args ...string) {
	assert.Len(t, job.Spec.Containers, 1)
	assert.Equal(t, img, job.Spec.Containers[0].Image)
	assert.Len(t, job.Spec.Containers[0].Command, 1)
	//assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh", job.Spec.Containers[0].Command[0])
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
func AssertThatContainerHasEnv(t *testing.T, cont v1.Container, expName, expValue string) {
	for _, env := range cont.Env {
		if env.Name == expName && env.Value == expValue {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Container [%s] does not have environment variable [%s] with value [%s]", cont.Name, expName, expValue))
}

// AssertThatContainerHasEnvFromSecret checks if container has specified given environment variable
func AssertThatContainerHasEnvFromSecret(t *testing.T, cont v1.Container, expName, expSecretName, expSecretKey string) {
	for _, env := range cont.Env {
		if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == expSecretName && env.ValueFrom.SecretKeyRef.Key == expSecretKey {
			return
		}
	}
	assert.Fail(t, fmt.Sprintf("Container [%s] does not have environment variable [%s] with value from secret [name: %s, key: %s]", cont.Name, expName, expSecretName, expSecretKey))
}
