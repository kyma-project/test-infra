package controlplane_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKCPGKEProvisionerIntegrationPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/control-plane/control-plane-gke-provisioner-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/control-plane"}), "pre-main-control-plane-gke-provisioner-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Optional)

	assert.Equal(t, "^resources/kcp/charts/provisioner/", actualJob.RunIfChanged)
	assert.Equal(t, "github.com/kyma-project/control-plane", actualJob.PathAlias)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	tester.AssertThatHasPresets(t, actualJob.JobBase,
		preset.KymaGuardBotGithubToken,
		preset.KymaKeyring,
		preset.KymaEncriptionKey,
		"preset-kms-gc-project-env",
		preset.SaGKEKymaIntegration,
		"preset-gc-compute-envs",
		preset.GCProjectEnv,
		"preset-docker-push-repository-gke-integration",
		preset.DindEnabled,
		"preset-kyma-artifacts-bucket",
		preset.GardenerAzureIntegration,
		preset.BuildPr,
		"preset-kyma-development-artifacts-bucket",
		preset.ClusterVersion,
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "main")
	require.Len(t, actualJob.Spec.Containers, 1)
	kcpCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, kcpCont.Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/control-plane-gke-integration.sh"}, kcpCont.Command)
	require.Len(t, kcpCont.Args, 0)
	tester.AssertThatContainerHasEnv(t, kcpCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, kcpCont, "RUN_PROVISIONER_TESTS", "true")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}
