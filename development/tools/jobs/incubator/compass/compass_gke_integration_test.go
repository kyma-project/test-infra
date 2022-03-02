package compass_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompassGKEIntegrationPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/compass"}), "pre-main-compass-gke-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.False(t, actualJob.Optional)

	assert.Equal(t, "^((chart\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
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
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "main")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/compass-gke-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	tester.AssertThatContainerHasEnv(t, compassCont, "GKE_CLUSTER_VERSION", "1.20")
}

func TestCompassGKEIntegrationJobsReleases(t *testing.T) {
	for _, currentRelease := range releases.GetAllKymaReleases() {
		t.Run(currentRelease.String(), func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/compass"}), tester.GetReleaseJobName("compass-gke-integration", currentRelease), currentRelease.Branch())
			require.NotNil(t, actualPresubmit)
			assert.True(t, actualPresubmit.Optional)
			assert.False(t, actualPresubmit.SkipReport)

			assert.Equal(t, "github.com/kyma-incubator/compass", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease.Branch())
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase,
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
				preset.BuildRelease,
				"preset-kyma-development-artifacts-bucket",
			)
			assert.False(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.Spec.Containers, 1)
			testContainer := actualPresubmit.Spec.Containers[0]

			assert.Equal(t, tester.ImageKymaIntegrationLatest, testContainer.Image)

			assert.Len(t, testContainer.Command, 1)
			tester.AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
		})
	}
}

func TestCompassGKEIntegrationPostsubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/compass"}), "post-main-compass-gke-integration", "master")
	require.NotNil(t, actualJob)

	// then

	assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
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
		preset.BuildMaster,
		"preset-kyma-development-artifacts-bucket",
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "main")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/compass-gke-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	tester.AssertThatContainerHasEnv(t, compassCont, "GKE_CLUSTER_VERSION", "1.20")
}
