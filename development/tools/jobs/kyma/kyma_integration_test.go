package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaIntegrationJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	kymaPresubmits, ex := jobConfig.Presubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 2)

	actualVM := kymaPresubmits[0]
	assert.Equal(t, "kyma-integration", actualVM.Name)
	assert.Equal(t, "^(resources|installation)", actualVM.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, actualVM, "resources/values.yaml")
	tester.AssertThatJobRunIfChanged(t, actualVM, "installation/file.yaml")
	assert.True(t, actualVM.SkipReport)
	assert.Equal(t, 10, actualVM.MaxConcurrency)
	// TODO add assertions about presets
	assert.Equal(t, []string{"master"}, actualVM.Branches)
	assert.True(t, actualVM.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualVM.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig)
	assert.Equal(t, tester.ImageBoostrap001, actualVM.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma.sh"}, actualVM.Spec.Containers[0].Command)

	actualGKE := kymaPresubmits[1]
	assert.Equal(t, "kyma-gke-integration", actualGKE.Name)
	assert.Equal(t, "^(resources|installation)", actualGKE.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, actualGKE, "resources/values.yaml")
	tester.AssertThatJobRunIfChanged(t, actualGKE, "installation/file.yaml")
	assert.True(t, actualGKE.SkipReport)
	assert.Equal(t, 10, actualGKE.MaxConcurrency)
	// TODO add assertions about presets
	assert.Equal(t, []string{"master"}, actualGKE.Branches)
	assert.True(t, actualGKE.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualGKE.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig)
	assert.Equal(t, tester.ImageBootstrapHelm20181121, actualGKE.Spec.Containers[0].Image)
}

func TestKymaIntegrationJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	kymaPostsubmits, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPostsubmits, 2)

	actualVM := kymaPostsubmits[0]
	assert.Equal(t, "kyma-integration", actualVM.Name)
	assert.Equal(t, []string{"master"}, actualVM.Branches)
	assert.Equal(t, 10, actualVM.MaxConcurrency)
	assert.Equal(t, "", actualVM.RunIfChanged)
	assert.True(t, actualVM.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualVM.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig)
	// TODO add assertions about presets
	assert.Equal(t, tester.ImageBoostrap001, actualVM.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma.sh"}, actualVM.Spec.Containers[0].Command)

	actualGKE := kymaPostsubmits[1]
	assert.Equal(t, "kyma-gke-integration", actualGKE.Name)
	assert.Equal(t, "", actualGKE.RunIfChanged)
	assert.Equal(t, 10, actualGKE.MaxConcurrency)
	// TODO add assertions about presets
	assert.Equal(t, []string{"master"}, actualGKE.Branches)
	assert.True(t, actualGKE.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualGKE.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig)

	assert.Equal(t, tester.ImageBootstrapHelm20181121, actualGKE.Spec.Containers[0].Image)
}
