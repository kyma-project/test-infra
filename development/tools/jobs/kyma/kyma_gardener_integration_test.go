package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaGardenerAzureIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener-azure.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	jobName := "kyma-integration-gardener-azure"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)
	assert.True(t, job.Decorate)
	assert.Equal(t, "0 4,7,10,13 * * *", job.Cron)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable)
	tester.AssertThatHasExtraRefs(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, tester.ImageKymaIntegrationK15, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener-azure.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "westeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerGCPIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener-gcp.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	jobName := "kyma-integration-gardener-gcp"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)
	assert.True(t, job.Decorate)
	assert.Equal(t, "00 08 * * *", job.Cron)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerGCPIntegration, preset.KymaCLIStable)
	tester.AssertThatHasExtraRefs(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, tester.ImageKymaIntegrationK15, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener-gcp.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "europe-west4")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerAWSIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener-aws.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	jobName := "kyma-integration-gardener-aws"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)
	assert.True(t, job.Decorate)
	assert.Equal(t, "00 14 * * *", job.Cron)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAWSIntegration, preset.KymaCLIStable)
	tester.AssertThatHasExtraRefs(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20200206-22eb97a4", job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener-aws.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "eu-west-3")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "eu-west-3a")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
