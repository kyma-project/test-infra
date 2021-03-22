package terraform_provider_gardener_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformProviderGardenerJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/terraform-provider-gardener/terraform-provider-gardener.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/terraform-provider-gardener"})

	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(kymaPresubmits, "pre-master-terraform-provider-gardener", "master")
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Empty(t, actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled)
	assert.Equal(t, tester.ImageGolangBuildpack1_14, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, "GO111MODULE", actualPresubmit.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "on", actualPresubmit.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/terraform-provider-gardener", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestTerraformProviderGardenerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/terraform-provider-gardener/terraform-provider-gardener.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/terraform-provider-gardener"})
	assert.Len(t, kymaPost, 1)

	actualPost := tester.FindPostsubmitJobByNameAndBranch(kymaPost, "post-master-terraform-provider-gardener", "master")
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled)
	assert.Equal(t, tester.ImageGolangBuildpack1_14, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, "GO111MODULE", actualPost.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "on", actualPost.Spec.Containers[0].Env[0].Value)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/terraform-provider-gardener", "ci-master"}, actualPost.Spec.Containers[0].Args)
}
