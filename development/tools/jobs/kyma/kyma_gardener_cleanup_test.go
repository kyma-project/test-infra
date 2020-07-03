package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaGardenerCleanupJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-gardener-cleanup.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 1)

	jobName := "kyma-gardener-cleanup"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)
	assert.True(t, job.Decorate)
	assert.Equal(t, "0 * * * *", job.Cron)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerGCPIntegration)
	tester.AssertThatHasExtraRepoRef(t, job.JobBase.UtilityConfig, []string{"test-infra"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/helpers/cleanup-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
