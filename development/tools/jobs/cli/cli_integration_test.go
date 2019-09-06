package kymacli_test

import (
	"testing"

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
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetBuildPr, tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-cli.sh"}, actualPresubmit.Spec.Containers[0].Command)
}

func TestKymaCliIntegrationJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(cliIntegrationJobPath)
	// THEN
	require.NoError(t, err)

	expName := "post-kyma-cli-integration"
	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/cli"], expName, "master")
	require.NotNil(t, actualPost)

	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/cli", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBuildMaster, tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-cli.sh"}, actualPost.Spec.Containers[0].Command)
}
