package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const runtimeAgentTestJobPath = "./../../../../prow/jobs/kyma/tests/compass-runtime-agent/compass-runtime-agent-test.yaml"

func TestRuntimeAgentTestJobReleases(t *testing.T) {
	// WHEN
	unsupportedReleases := []tester.SupportedRelease{tester.Release11, tester.Release12, tester.Release13}

	for _, currentRelease := range tester.GetKymaReleaseBranchesBesides(unsupportedReleases) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig(runtimeAgentTestJobPath)
			// THEN
			require.NoError(t, err)

			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-tests-compass-runtime-agent", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPresubmit.AlwaysRun)
			tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/kyma/tests/compass-runtime-agent")
		})
	}
}

func TestRuntimeAgentTestJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(runtimeAgentTestJobPath)
	// THEN
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-tests-compass-runtime-agent", "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)

	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^tests/compass-runtime-agent/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "tests/compass-runtime-agent/some_random_file.go")
	tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/kyma/tests/compass-runtime-agent")
}

func TestRuntimeAgentTestJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(runtimeAgentTestJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], "post-master-kyma-tests-compass-runtime-agent", "master")
	require.NotNil(t, actualPost)

	expName := "post-master-kyma-tests-compass-runtime-agent"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^tests/compass-runtime-agent/", actualPost.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPost, "tests/compass-runtime-agent/some_random_file.go")
	tester.AssertThatExecGolangBuildpack(t, actualPost.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/kyma/tests/compass-runtime-agent")
}
