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
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")
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
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")
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
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")
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
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")

	assert.Equal(t, tester.ImageBootstrapHelm20181121, actualGKE.Spec.Containers[0].Image)
}

func TestKymaIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 4)

	expName := "orphaned-disks-cleaner"
	disksCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, disksCleanerPeriodic)
	assert.Equal(t, expName, disksCleanerPeriodic.Name)
	assert.True(t, disksCleanerPeriodic.Decorate)
	assert.Equal(t, "*/10 * * * *", disksCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, disksCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, disksCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", disksCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, disksCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/disks-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=true"}, disksCleanerPeriodic.Spec.Containers[0].Args)

	expName = "orphaned-clusters-cleaner"
	clustersCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, clustersCleanerPeriodic)
	assert.Equal(t, expName, clustersCleanerPeriodic.Name)
	assert.True(t, clustersCleanerPeriodic.Decorate)
	assert.Equal(t, "*/10 * * * *", clustersCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, clustersCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, clustersCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", clustersCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, clustersCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/clusters-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=true"}, clustersCleanerPeriodic.Spec.Containers[0].Args)

	expName = "orphaned-loadbalancer-cleaner"
	loadbalancerCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, loadbalancerCleanerPeriodic)
	assert.Equal(t, expName, loadbalancerCleanerPeriodic.Name)
	assert.True(t, loadbalancerCleanerPeriodic.Decorate)
	assert.Equal(t, "*/10 * * * *", loadbalancerCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, loadbalancerCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, loadbalancerCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", loadbalancerCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/loadbalancer-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=true"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Args)
}
