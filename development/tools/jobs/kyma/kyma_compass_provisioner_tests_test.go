package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaGKECompassProvisionerTestsPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-compass-provisioner-tests.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-gke-compass-provisioner-tests", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Optional)
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "resources/compass/charts/provisioner", actualJob.RunIfChanged)
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	assert.False(t, actualJob.SkipReport)
	tester.AssertThatHasPresets(t, actualJob.JobBase,
		"preset-kyma-guard-bot-github-token",
		"preset-kyma-keyring",
		"preset-kyma-encryption-key",
		"preset-kms-gc-project-env",
		"preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs",
		"preset-gc-project-env",
		"preset-docker-push-repository-gke-integration",
		"preset-dind-enabled",
		"preset-kyma-artifacts-bucket",
		"preset-gardener-azure-kyma-integration",
		"preset-build-pr",
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationK15, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, compassCont, "RUN_PROVISIONER_TESTS", "true")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestKymaGKECompassProvisionerTestsPostsubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-compass-provisioner-tests.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-project/kyma"], "post-master-kyma-gke-compass-provisioner-tests", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	tester.AssertThatHasPresets(t, actualJob.JobBase,
		"preset-kyma-guard-bot-github-token",
		"preset-kyma-keyring",
		"preset-kyma-encryption-key",
		"preset-kms-gc-project-env",
		"preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs",
		"preset-gc-project-env",
		"preset-docker-push-repository-gke-integration",
		"preset-dind-enabled",
		"preset-kyma-artifacts-bucket",
		"preset-gardener-azure-kyma-integration",
		"preset-build-master",
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationK15, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, compassCont, "RUN_PROVISIONER_TESTS", "true")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}
