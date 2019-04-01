package kyma_test

import (
	"testing"

	"k8s.io/test-infra/prow/config"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaWebsiteBuildTriggerJobReleases(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetKymaReleaseBranchesBesides([]string{"release-0.7", "release-0.8", "release-0.9"}) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/docs/.ci/website-build-trigger.yaml")
			// THEN
			require.NoError(t, err)
			actualPostsubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], tester.GetReleasePostSubmitJobName("kyma-website-build-trigger", currentRelease), currentRelease)
			require.NotNil(t, actualPostsubmit)
			assert.True(t, actualPostsubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPostsubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, tester.PresetBotGithubToken, tester.PresetBotGithubSSH, tester.PresetBotGithubIdentity, tester.PresetBuildRelease)
			assertWebsiteBuildTriggerJob(t, actualPostsubmit)
		})
	}
}

func TestKymaWebsiteBuildTriggerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/docs/.ci/website-build-trigger.yaml")
	// THEN
	require.NoError(t, err)

	kymaPost, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)

	expName := "post-master-kyma-website-build-trigger"
	actualPost := tester.FindPostsubmitJobByName(kymaPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBotGithubToken, tester.PresetBotGithubSSH, tester.PresetBotGithubIdentity, tester.PresetBuildMaster)
	tester.AssertThatHasExtraRefs(t, actualPost.JobBase.UtilityConfig, []string{"test-infra", "website"})
	assert.Equal(t, "^docs/", actualPost.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPost, "docs/some_random_file.go")
	assertWebsiteBuildTriggerJob(t, actualPost)
}

func assertWebsiteBuildTriggerJob(t *testing.T, postsubmit *config.Postsubmit) {
	assert.Len(t, postsubmit.Spec.Containers, 1)
	assert.Equal(t, tester.ImageGolangBuildpack1_11, postsubmit.Spec.Containers[0].Image)
	assert.Len(t, postsubmit.Spec.Containers[0].Command, 1)
	assert.Equal(t, "bash", postsubmit.Spec.Containers[0].Command[0])
	assert.Equal(t, []string{
		"-c",
		"GITHUB_TOKEN=${BOT_GITHUB_TOKEN} /home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/job-guard.sh && cd /home/prow/go/src/github.com/kyma-project/kyma/docs/.ci && make ci-release",
	}, postsubmit.Spec.Containers[0].Args)
}
