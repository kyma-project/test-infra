package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaGKECompassIntegrationOnDemand(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/compass-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-gke-compass-integration-on-demand", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.True(t, actualJob.Optional)
	assert.Empty(t, actualJob.RunIfChanged)
	tester.AssertThatHasPresets(t, actualJob.JobBase,
		"prow.kyma-project.io/slack.skipReport",
		"preset-kyma-guard-bot-github-token",
		"preset-kyma-keyring",
		"preset-kyma-encryption-key",
		"preset-kms-gc-project-env",
		"preset-build-master",
		"preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs",
		"preset-gc-project-env",
		"preset-docker-push-repository-gke-integration",
		"preset-dind-enabled",
		"preset-kyma-artifacts-bucket",
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190528-8897828", compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, compassCont, "INPUT_CLUSTER_NAME", "compass-integration-periodic-on-demand")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestCompassIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/compass-integration.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	expName := "kyma-gke-compass-integration-periodic"
	compassPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, compassPeriodic)
	tester.AssertThatHasPresets(t, compassPeriodic.JobBase,
		"preset-kyma-keyring",
		"preset-kyma-encryption-key",
		"preset-kyma-guard-bot-github-token",
		"preset-kms-gc-project-env",
		"preset-build-master",
		"preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs",
		"preset-gc-project-env",
		"preset-docker-push-repository-gke-integration",
		"preset-dind-enabled",
		"preset-kyma-artifacts-bucket",
	)
	tester.AssertThatHasExtraRefTestInfra(t, compassPeriodic.JobBase.UtilityConfig, "master")
	tester.AssertThatHasExtraRefs(t, compassPeriodic.JobBase.UtilityConfig, []string{"kyma"})
	require.Len(t, compassPeriodic.Spec.Containers, 1)
	compassCont := compassPeriodic.Spec.Containers[0]
	assert.True(t, compassPeriodic.Decorate)
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190528-8897828", compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, compassCont, "INPUT_CLUSTER_NAME", "compass-integration-periodic")
}
