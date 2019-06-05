package marketplaces_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const registryJobPath = "./../../../../../prow/jobs/incubator/marketplaces/marketplaces.yaml"

func TestMarketplacesJobRelease(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/marketplaces"], "rel-marketplaces", "1.1.1")
	require.NotNil(t, actualPost)
	actualPost = tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/marketplaces"], "rel-marketplaces", "2.1.1-rc1")
	require.NotNil(t, actualPost)

	assert.True(t, actualPost.Decorate)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoMarketplaces, tester.PresetGcrPush, tester.PresetBuildRelease, tester.PresetBotGithubToken)

}

func TestMarketplacesJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	actualPre := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-incubator/marketplaces"], "pre-marketplaces", "master")
	require.NotNil(t, actualPre)

	assert.Equal(t, 10, actualPre.MaxConcurrency)
	assert.False(t, actualPre.SkipReport)
	assert.True(t, actualPre.Decorate)
	assert.False(t, actualPre.Optional)
	assert.True(t, actualPre.AlwaysRun)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPre.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPre.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPre.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoMarketplaces, tester.PresetGcrPush, tester.PresetBuildPr)
}

func TestMarketplacesPostsubmit(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/marketplaces"], "post-marketplaces", "master")
	require.NotNil(t, actualPost)
	actualPost = tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/marketplaces"], "post-marketplaces", "release-1.1")
	require.NotNil(t, actualPost)

	assert.True(t, actualPost.Decorate)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoMarketplaces, tester.PresetGcrPush, tester.PresetBuildMaster)
}
