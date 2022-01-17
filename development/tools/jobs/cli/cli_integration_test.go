package kymacli_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cliIntegrationJobPath = "./../../../../prow/jobs/cli/cli-integration.yaml"

func TestKymaCliIntegrationPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(cliIntegrationJobPath)
	// THEN
	require.NoError(t, err)

	expName := "pre-cli-integration-kyma-1"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/cli"}), expName, "main")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, "github.com/kyma-project/cli", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
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

	expName := "post-cli-integration-kyma-2"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/cli"}), expName, "main")
	require.NotNil(t, actualPost)

	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, 10, actualPost.MaxConcurrency)

	assert.Equal(t, "github.com/kyma-project/cli", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.BuildMaster, preset.GCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPost.Spec.Containers[0].Image)
	tester.AssertThatContainerHasEnv(t, actualPost.Spec.Containers[0], "GO111MODULE", "on")
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-cli.sh"}, actualPost.Spec.Containers[0].Command)
}
