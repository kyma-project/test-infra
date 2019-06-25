package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaGithubReleaseJobPostsubmit(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-github-release.yaml")
			// THEN
			require.NoError(t, err)
			actualPostsubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], tester.GetReleasePostSubmitJobName("kyma-github-release", currentRelease), currentRelease)
			require.NotNil(t, actualPostsubmit)

			assert.True(t, actualPostsubmit.Decorate)
			tester.AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, "preset-sa-kyma-artifacts", "preset-bot-github-token")
			assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPostsubmit.Spec.Containers[0].Image)
			assert.Equal(t, []string{"-c", "/home/prow/go/src/github.com/kyma-project/test-infra/development/github-release.sh -targetCommit=${RELEASE_TARGET_COMMIT} -githubRepoOwner=${REPO_OWNER} -githubRepoName=${REPO_NAME} -githubAccessToken=${BOT_GITHUB_TOKEN} -releaseVersionFilePath=/home/prow/go/src/github.com/kyma-project/test-infra/prow/RELEASE_VERSION"}, actualPostsubmit.Spec.Containers[0].Args)
		})
	}
}
