package testinfra

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchPodsReleases(t *testing.T) {

	for _, currentRelease := range releases.GetAllKymaReleases() {
		t.Run(currentRelease.String(), func(t *testing.T) {
			expectedImage := tester.ImageGolangBuildpack1_11
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/watch-pods.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/test-infra"], tester.GetReleaseJobName("test-infra-watch-pods", currentRelease), currentRelease.Branch())
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoKyma, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPresubmit.AlwaysRun)
			tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, expectedImage, "/home/prow/go/src/github.com/kyma-project/test-infra/watch-pods")
		})
	}
}

func TestWatchPodsJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/watch-pods.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/test-infra"], "pre-master-test-infra-watch-pods", "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoKyma, tester.PresetGcrPush, tester.PresetBuildPr)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "watch-pods/fix")
	assert.Equal(t, "^watch-pods/", actualPresubmit.RunIfChanged)
	tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/test-infra/watch-pods")
}

func TestWatchPodsJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/watch-pods.yaml")
	// THEN
	require.NoError(t, err)

	expName := "post-master-test-infra-watch-pods"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-project/test-infra"], expName, "master")
	require.NotNil(t, actualPost)

	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoKyma, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^watch-pods/", actualPost.RunIfChanged)
	tester.AssertThatExecGolangBuildpack(t, actualPost.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/test-infra/watch-pods")
}
