package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetControllerReleases(t *testing.T) {
	// WHEN
	var unsupportedReleases []string

	for _, currentRelease := range tester.GetSupportedReleases(unsupportedReleases) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/asset-store-controller-manager/asset-store-controller-manager.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-components-asset-store-controller-manager", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPresubmit.AlwaysRun)

			var args []string
			if tester.HasOneOfSuffixes(currentRelease, "-0.7", "-0.8") {
				args = []string{
					"/home/prow/go/src/github.com/kyma-project/kyma/components/assetstore-controller-manager",
				}
			} else {
				args = []string{
					"/home/prow/go/src/github.com/kyma-project/kyma/components/asset-store-controller-manager",
				}
			}

			tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangKubebuilderBuildpackLatest, args...)
		})
	}
}

func TestAssetControllerJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/asset-store-controller-manager/asset-store-controller-manager.yaml")
	// THEN
	require.NoError(t, err)

	expName := "pre-master-kyma-components-asset-store-controller-manager"
	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)

	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^components/asset-store-controller-manager/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "components/asset-store-controller-manager/lets_play.go")
	tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangKubebuilderBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/components/asset-store-controller-manager")
}

func TestAssetControllerPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/asset-store-controller-manager/asset-store-controller-manager.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	expName := "post-master-kyma-components-asset-store-controller-manager"
	actualPost := tester.FindPostsubmitJobByName(kymaPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^components/asset-store-controller-manager/", actualPost.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPost, "components/asset-store-controller-manager/coco_jambo_i_do_przodu.go")
	tester.AssertThatExecGolangBuildpack(t, actualPost.JobBase, tester.ImageGolangKubebuilderBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/components/asset-store-controller-manager")
}
