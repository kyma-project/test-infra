package marketplaces_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const registryJobPath = "./../../../../../prow/jobs/incubator/marketplaces/marketplaces.yaml"

func TestMarketplacesJobRelease(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/marketplaces"}), "rel-marketplaces", "1.1.1")
	require.NotNil(t, actualPost)
	actualPost = tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/marketplaces"}), "rel-marketplaces", "2.1.1-rc1")
	require.NotNil(t, actualPost)

	assert.True(t, actualPost.Decorate)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoGlobal, preset.GcrPush, preset.BuildRelease, preset.BotGithubToken)

}

func TestMarketplacesJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	actualPre := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/marketplaces"}), "pre-marketplaces", "master")
	require.NotNil(t, actualPre)

	assert.Equal(t, 10, actualPre.MaxConcurrency)
	assert.False(t, actualPre.SkipReport)
	assert.True(t, actualPre.Decorate)
	assert.False(t, actualPre.Optional)
	assert.True(t, actualPre.AlwaysRun)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPre.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPre.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPre.JobBase, preset.DindEnabled, preset.DockerPushRepoGlobal, preset.GcrPush, preset.BuildPr)
}

func TestMarketplacesPostsubmit(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/marketplaces"}), "post-marketplaces", "master")
	require.NotNil(t, actualPost)
	actualPost = tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/marketplaces"}), "post-marketplaces", "release-1.1")
	require.NotNil(t, actualPost)

	assert.True(t, actualPost.Decorate)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoGlobal, preset.GcrPush, preset.BuildMaster)
}

func TestGovernanceJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	expName := "pre-marketplaces-governance"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/marketplaces"}), expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Empty(t, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/marketplaces", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr, preset.DindEnabled)
	assert.Equal(t, "milv.config.yaml|.md$", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "milv.config.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "some_markdown.md"))
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--repository", "marketplaces", "--repository-org", "kyma-incubator"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestGovernanceJobPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(registryJobPath)
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 1)

	expName := "marketplaces-governance-nightly"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "0 3 * * 1-5", actualPeriodic.Cron)
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.DindEnabled)
	tester.AssertThatHasExtraRepoRef(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra"})
	tester.AssertThatHasExtraRef(t, actualPeriodic.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-incubator",
			Repo:      "marketplaces",
			BaseRef:   "master",
			PathAlias: "github.com/kyma-incubator/marketplaces",
		},
	})
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPeriodic.Spec.Containers[0].Command)
	repositoryDirArg := fmt.Sprintf("%s/marketplaces", tester.KymaIncubatorDir)
	assert.Equal(t, []string{"--repository", "marketplaces", "--repository-org", "kyma-incubator", "--repository-dir", repositoryDirArg, "--full-validation", "true"}, actualPeriodic.Spec.Containers[0].Args)
}
