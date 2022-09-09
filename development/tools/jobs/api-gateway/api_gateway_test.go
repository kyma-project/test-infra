package testinfra_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const apiGatewayJobPath = "./../../../../prow/jobs/api-gateway/api-gateway-build.yaml"

func TestApiGatewayJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(apiGatewayJobPath)

	// then
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/api-gateway"}), "pre-kyma-project-api-gateway", "main")
	require.NotNil(t, actualPresubmit)
	actualPresubmit = tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/api-gateway"}), "pre-kyma-project-api-gateway", "release-1.0")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

	assert.True(t, actualPresubmit.AlwaysRun)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/api-gateway", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestApiGatewayJobPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(apiGatewayJobPath)

	// then
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/api-gateway"}), "post-kyma-project-api-gateway", "main")
	require.NotNil(t, actualPostsubmit)
	actualPostsubmit = tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/api-gateway"}), "post-kyma-project-api-gateway", "release-1.0")
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)

	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/api-gateway", "ci-main"}, actualPostsubmit.Spec.Containers[0].Args)
}

func TestApiGatewayJobRelease(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(apiGatewayJobPath)

	// then
	require.NoError(t, err)

	actualRelease := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/api-gateway"}), "rel-api-gateway", "1.0.0")
	require.NotNil(t, actualRelease)

	assert.Equal(t, 10, actualRelease.MaxConcurrency)

	tester.AssertThatHasPresets(t, actualRelease.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualRelease.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualRelease.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/api-gateway", "ci-release"}, actualRelease.Spec.Containers[0].Args)
}
