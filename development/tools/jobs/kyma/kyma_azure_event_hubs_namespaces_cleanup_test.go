package kyma_test

import (
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaAzureEventhubsNamespacesCleanupJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-azure-event-hubs-namespaces-cleanup.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 1)

	jobName := "kyma-azure-event-hubs-namespaces-cleanup"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)
	assert.True(t, job.Decorate)
	assert.Equal(t, "30 * * * *", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 30*time.Minute)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasExtraRepoRef(t, job.JobBase.UtilityConfig, []string{"test-infra"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, "true", job.JobBase.Labels["preset-gardener-azure-kyma-integration"])
	assert.Equal(t, []string{"-c", "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/helpers/cleanup-azure-event-hubs-namespaces.sh"}, job.Spec.Containers[0].Args)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "AZURE_SUBSCRIPTION_NAME", "sap-se-cx-kyma-prow-dev")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "TTL_HOURS", "6")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
