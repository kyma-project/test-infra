package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestKymaIntegrationVMJobPostsubmit(t *testing.T) {
// 	// WHEN
// 	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
// 	// THEN
// 	require.NoError(t, err)

// 	kymaPostsubmits, ex := jobConfig.Postsubmits["kyma-project/kyma"]
// 	assert.True(t, ex)
// 	assert.Len(t, kymaPostsubmits, 2)

// 	actualVM := kymaPostsubmits[0]
// 	assert.Equal(t, "kyma-integration", actualVM.Name)
// 	assert.Equal(t, []string{"master"}, actualVM.Branches)
// 	assert.Equal(t, 10, actualVM.MaxConcurrency)
// 	assert.Equal(t, "", actualVM.RunIfChanged)
// 	assert.True(t, actualVM.Decorate)
// 	assert.Equal(t, "github.com/kyma-project/kyma", actualVM.PathAlias)
// 	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")
// 	tester.AssertThatHasPresets(t, actualVM.JobBase, tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration")
// 	assert.Equal(t, tester.ImageBootstrap001, actualVM.Spec.Containers[0].Image)
// 	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma.sh"}, actualVM.Spec.Containers[0].Command)
// 	tester.AssertThatSpecifiesResourceRequests(t, actualVM.JobBase)
// }

// func TestKymaGithubReleaseJobPostsubmit(t *testing.T) {
// 	// WHEN
// 	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-github-release.yaml")
// 	// THEN
// 	require.NoError(t, err)

// 	kymaPostsubmits, ex := jobConfig.Postsubmits["kyma-project/kyma"]
// 	assert.True(t, ex)
// 	assert.Len(t, kymaPostsubmits, 1)

// 	actualPostsubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-components-binding-usage-controller", currentRelease)
// 	assert.Equal(t, "kyma-github-release", actual.Name)
// 	assert.Equal(t, "", actual.RunIfChanged)
// 	assert.Equal(t, 1, actual.MaxConcurrency)
// 	tester.AssertThatHasPresets(t, actual.JobBase, "preset-sa-kyma-artifacts", "preset-bot-github-token")
// 	assert.Equal(t, []string{"master"}, actual.Branches)
// 	// assert.True(t, actualGKE.Decorate)
// 	// assert.Equal(t, "github.com/kyma-project/kyma", actualGKE.PathAlias)
// 	tester.AssertThatHasExtraRefTestInfra(t, actual.JobBase.UtilityConfig, "release-0.6")
// 	assert.Equal(t, tester.ImageGolangBuildpackLatest, actual.Spec.Containers[0].Image)
// 	tester.AssertThatSpecifiesResourceRequests(t, actual.JobBase)
// }

func TestKymaGithubReleaseJobPostsubmit(t *testing.T) {
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-github-release.yaml")
			// THEN
			require.NoError(t, err)
			actualPostsubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], "kyma-github-release", currentRelease)
			require.NotNil(t, actualPostsubmit)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPostsubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, "preset-sa-kyma-artifacts", "preset-bot-github-token")
			assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPostsubmit.Spec.Containers[0].Image)
		})
	}
}
