package compass_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompassGKEIntegrationPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/compass"}), "pre-master-compass-gke-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Optional)
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "^((chart\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
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
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/compass-gke-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestCompassGKEIntegrationJobsReleases(t *testing.T) {
	for _, currentRelease := range releases.GetAllKymaReleases() {
		if currentRelease.IsNotOlderThan(releases.Release113) {
			t.Run(currentRelease.String(), func(t *testing.T) {
				jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
				// THEN
				require.NoError(t, err)
				actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/compass"}), tester.GetReleaseJobName("compass-gke-integration", currentRelease), currentRelease.Branch())
				require.NotNil(t, actualPresubmit)
				assert.True(t, actualPresubmit.Optional)
				assert.False(t, actualPresubmit.SkipReport)
				assert.True(t, actualPresubmit.Decorate)
				assert.Equal(t, "github.com/kyma-incubator/compass", actualPresubmit.PathAlias)
				tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease.Branch())
				tester.AssertThatHasPresets(t, actualPresubmit.JobBase,
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
					"preset-build-release",
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
}

func TestCompassGKEIntegrationPostsubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/compass"}), "post-master-compass-gke-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
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
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationLatest, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/compass-gke-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}

func TestCompassGKEIntegrationPeriodic(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/compass/compass-gke-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPeriodicJobByName(jobConfig.AllPeriodics(), "kyma-compass-gke-integration-daily")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	assert.Equal(t, "00 00 * * *", actualJob.Cron)
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
	compassCont := actualJob.Spec.Containers[0]
	assert.Equal(t, tester.ImageKymaIntegrationK15, compassCont.Image)
	assert.Equal(t, []string{"bash"}, compassCont.Command)
	require.Len(t, compassCont.Args, 2)
	assert.Equal(t, "-c", compassCont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/compass-gke-integration.sh", compassCont.Args[1])
	tester.AssertThatContainerHasEnv(t, compassCont, "CLOUDSDK_COMPUTE_ZONE", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
}
