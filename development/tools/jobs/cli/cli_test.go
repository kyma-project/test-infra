package kymacli_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const registryJobPath = "./../../../../prow/jobs/cli/cli.yaml"

func TestKymaCliJobRelease(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	expName := "rel-kyma-cli"
	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], expName, "1.1.1")
	require.NotNil(t, actualPost)
	actualPost = tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], expName, "2.1.1-rc1")
	require.NotNil(t, actualPost)

	assert.True(t, actualPost.Decorate)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.Equal(t, "github.com/kyma-project/cli", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBuildRelease, tester.PresetBotGithubToken)
}

func TestKymaCliJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	expName := "pre-kyma-cli"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, "github.com/kyma-project/cli", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetBuildPr)
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/cli"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestKymaCliJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	expName := "post-kyma-cli"
	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPost)
	actualPost = tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], expName, "release-1.1")
	require.NotNil(t, actualPost)

	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/cli", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/cli"}, actualPost.Spec.Containers[0].Args)
}
