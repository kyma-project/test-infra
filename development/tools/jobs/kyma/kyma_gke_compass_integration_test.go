package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaGKECompassIntegrationPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-gke-compass-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"}), "pre-master-kyma-gke-compass-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "^((resources\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
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
		preset.ClusterVersion,
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestKymaGKECompassIntegrationJobsReleases(t *testing.T) {
	for _, currentRelease := range releases.GetAllKymaReleases() {
		t.Run(currentRelease.String(), func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-gke-compass-integration.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"}), tester.GetReleaseJobName("kyma-gke-compass-integration", currentRelease), currentRelease.Branch())
			require.NotNil(t, actualPresubmit)
			assert.True(t, actualPresubmit.Optional)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
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
				preset.ClusterVersion,
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

func TestKymaGKECompassIntegrationPostsubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-gke-compass-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/kyma"}), "post-master-kyma-gke-compass-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
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
		preset.ClusterVersion,
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestCompassGKEIntegrationPeriodic(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-gke-compass-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPeriodicJobByName(jobConfig.AllPeriodics(), "kyma-gke-compass-integration-daily")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	assert.Equal(t, "00 00 * * *", actualJob.Cron)
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
		preset.ClusterVersion,
	)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	require.Len(t, actualJob.Spec.Containers, 1)
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-compass-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}
