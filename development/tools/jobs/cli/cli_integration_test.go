package kymacli_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cliIntegrationJobPath = "./../../../../prow/jobs/cli/cli-integration.yaml"

func TestKymaCliIntegrationPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(cliIntegrationJobPath)
	// THEN
	require.NoError(t, err)

	expName := "pre-kyma-cli-integration"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, "github.com/kyma-project/cli", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr, preset.GCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	tester.AssertThatContainerHasEnv(t, actualPresubmit.Spec.Containers[0], "GO111MODULE", "on")
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-cli.sh"}, actualPresubmit.Spec.Containers[0].Command)
}

func TestKymaCliIntegrationJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(cliIntegrationJobPath)
	// THEN
	require.NoError(t, err)

	expName := "post-kyma-cli-integration"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPost)

	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/cli", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.BuildMaster, preset.GCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPost.Spec.Containers[0].Image)
	tester.AssertThatContainerHasEnv(t, actualPost.Spec.Containers[0], "GO111MODULE", "on")
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-cli.sh"}, actualPost.Spec.Containers[0].Command)
}

func TestKymaCliIntegrationGKEPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(cliIntegrationJobPath)
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	expName := "kyma-cli-integration-gke"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "00 00 * * *", actualPeriodic.Cron)
	tester.AssertThatHasExtraRefs(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra", "cli"})
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.SaGKEKymaIntegration, preset.GCProjectEnv, "preset-gc-compute-envs", "preset-cluster-use-ssd")
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPeriodic.Spec.Containers[0].Image)
	tester.AssertThatSpecifiesResourceRequests(t, actualPeriodic.JobBase)
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "CLOUDSDK_COMPUTE_ZONE", "europe-west4-a")
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "GO111MODULE", "on")
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-gke-integration-cli.sh"}, actualPeriodic.Spec.Containers[0].Command)
}
