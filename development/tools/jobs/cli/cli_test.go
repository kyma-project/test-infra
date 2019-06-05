package kymacli_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaCliReleases(t *testing.T) {
	// WHEN
	unsupportedReleases := []tester.SupportedRelease{tester.Release09, tester.Release10}

	for _, currentRelease := range tester.GetKymaReleaseBranchesBesides(unsupportedReleases) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/cli/cli.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/cli"], tester.GetReleaseJobName("kyma-cli", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/cli", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetBuildPr)
			assert.True(t, actualPresubmit.AlwaysRun)
			tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/cli")

			actualPostsubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], tester.GetReleasePostSubmitJobName("kyma-cli", currentRelease), currentRelease)
			require.NotNil(t, actualPostsubmit)
			assert.True(t, actualPostsubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/cli", actualPostsubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, tester.PresetBuildRelease)
			tester.AssertThatExecGolangBuildpack(t, actualPostsubmit.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/cli")
		})
	}
}

func TestKymaCliJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/cli/cli.yaml")
	// THEN
	require.NoError(t, err)

	expName := "pre-master-kyma-cli"
	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
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
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/cli/cli.yaml")
	// THEN
	require.NoError(t, err)

	expName := "post-master-kyma-cli"
	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/cli", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/cli"}, actualPost.Spec.Containers[0].Args)
}
