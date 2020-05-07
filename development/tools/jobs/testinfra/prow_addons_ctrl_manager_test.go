package testinfra_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const prowAddonsCtrlManagerJobPath = "./../../../../prow/jobs/test-infra/prow-addons-ctrl-manager.yaml"

func TestProwAddonsCtrlManagerJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(prowAddonsCtrlManagerJobPath)

	// then
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"}), "pre-master-test-infra-development-prow-addons-ctrl-manager", "master")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^development/prow-addons-ctrl-manager/", actualPresubmit.RunIfChanged)
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200124-69faeef6", actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestProwAddonsCtrlManagerJobPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(prowAddonsCtrlManagerJobPath)

	// then
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"}), "post-master-test-infra-development-prow-addons-ctrl-manager", "master")
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)
	assert.True(t, actualPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPostsubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildMaster)
	assert.Equal(t, "^development/prow-addons-ctrl-manager/", actualPostsubmit.RunIfChanged)
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/buildpack-golang-kubebuilder2:v20200124-69faeef6", actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager"}, actualPostsubmit.Spec.Containers[0].Args)
}
