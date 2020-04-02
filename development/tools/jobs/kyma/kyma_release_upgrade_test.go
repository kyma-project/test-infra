package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaReleaseUpgradeJobsPostsubmit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("", func(t *testing.T) { //todo: desc

		//given
		jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-release-upgrade.yaml")
		require.NoError(err)

		//when
		actualJob := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], "post-kyma-release-upgrade")

		//then
		require.NotNil(actualJob)
		assert.Equal("github.com/kyma-project/kyma", actualJob.PathAlias)
		assert.True(actualJob.Decorate)
		tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
		tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master") //todo: ???
		assert.Equal(tester.ImageKymaIntegrationK15, actualJob.Spec.Containers[0].Image)
		assert.Equal([]string{"-c", "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-gke-rel2rel-upgrade.sh"}, actualJob.Spec.Containers[0].Args)
		tester.AssertThatHasPresets(t, actualJob.JobBase, preset.DindEnabled, preset.BotGithubToken, preset.GCProjectEnv, "preset-gc-compute-envs")
		tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_COMPUTE_ZONE", "europe-west4-a")
	})
}
