package controlplane_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKCPGKEProvisionerIntegrationPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/control-plane/control-plane-gke-provisioner-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/control-plane"}), "pre-master-control-plane-gke-provisioner-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Optional)
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "resources/provisioner", actualJob.RunIfChanged)
	assert.Equal(t, "github.com/kyma-project/control-plane", actualJob.PathAlias)
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
		"preset-build-pr",
		"preset-kyma-development-artifacts-bucket",
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	kcpCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, kcpCont.Image)
	assert.Equal(t, []string{"bash"}, kcpCont.Command)
	require.Len(t, kcpCont.Args, 2)
	assert.Equal(t, "-c", kcpCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/control-plane-gke-integration.sh", kcpCont.Args[1])
	tester.AssertThatContainerHasEnv(t, kcpCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, kcpCont, "RUN_PROVISIONER_TEST", "true")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestKCPGKEProvisionerIntegrationPostsubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/control-plane/control-plane-gke-provisioner-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/control-plane"}), "post-master-control-plane-gke-provisioner-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "github.com/kyma-project/control-plane", actualJob.PathAlias)
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
		"preset-kyma-development-artifacts-bucket",
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	kcpCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, kcpCont.Image)
	assert.Equal(t, []string{"bash"}, kcpCont.Command)
	require.Len(t, kcpCont.Args, 2)
	assert.Equal(t, "-c", kcpCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/control-plane-gke-integration.sh", kcpCont.Args[1])
	tester.AssertThatContainerHasEnv(t, kcpCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, kcpCont, "RUN_PROVISIONER_TEST", "true")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}
