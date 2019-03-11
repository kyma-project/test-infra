package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubelessIntegrationTestsJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/kubeless-integration/kubeless-integration.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-tests-end-to-end-kubeless-integration", "master")

	expName := "pre-master-kyma-tests-end-to-end-kubeless-integration"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^tests/end-to-end/kubeless-integration/", actualPresubmit.RunIfChanged)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/kubeless-integration"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestKubelessIntegrationTestsJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/kubeless-integration/kubeless-integration.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-master-kyma-tests-end-to-end-kubeless-integration"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^tests/end-to-end/kubeless-integration/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/kubeless-integration"}, actualPost.Spec.Containers[0].Args)
}

func TestKubelessIntegrationReleases(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			// Retaining the behavior for release-0.6 and release-0.7
			if currentRelease == "release-0.6" || currentRelease == "release-0.7" {
				jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/kubeless-integration/kubeless-integration.yaml")
				// THEN
				require.NoError(t, err)
				actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-tests-kubeless-integration", currentRelease), currentRelease)
				require.NotNil(t, actualPresubmit)
				assert.False(t, actualPresubmit.SkipReport)
				assert.True(t, actualPresubmit.Decorate)
				assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
				tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
				tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
				assert.True(t, actualPresubmit.AlwaysRun)
				tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/tests/kubeless-integration")
			} else {
				jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/kubeless-integration/kubeless-integration.yaml")
				// THEN
				require.NoError(t, err)
				actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-tests-kubeless-integration", currentRelease), currentRelease)
				require.NotNil(t, actualPresubmit)
				assert.False(t, actualPresubmit.SkipReport)
				assert.True(t, actualPresubmit.Decorate)
				assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
				tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
				tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
				assert.True(t, actualPresubmit.AlwaysRun)
				tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/kubeless-integration")
			}

		})
	}
}
