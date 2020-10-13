package kyma_test

import (
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// add test here
func TestKymaGardenerAzureUpgradeJobPeriodics(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-upgrade-gardener.yaml")
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	jobName := "kyma-upgrade-gardener-azure"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)
	assert.True(t, job.Decorate)
	assert.Equal(t, "0 1 * * 1-5", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 4*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.KymaGuardBotGithubToken)
	tester.AssertThatHasExtraRepoRef(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-upgrade-gardener-azure.sh"}, job.Spec.Containers[0].Args)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "westeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
