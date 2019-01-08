package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaIntegrationVMJobsReleases(t *testing.T) {
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-integration", currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration")
			assert.False(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.Spec.Containers, 1)
			testContainer := actualPresubmit.Spec.Containers[0]
			assert.Equal(t, tester.ImageBootstrap001, testContainer.Image)
			assert.Len(t, testContainer.Command, 1)
			assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma.sh", testContainer.Command[0])
			tester.AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
		})
	}
}

func TestKymaIntegrationVMJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	actualVM := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-integration", "master")
	require.NotNil(t, actualVM)
	assert.Equal(t, "kyma-integration", actualVM.Name)
	assert.Equal(t, "^(resources|installation)", actualVM.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualVM, "resources/values.yaml")
	tester.AssertThatJobRunIfChanged(t, *actualVM, "installation/file.yaml")
	assert.False(t, actualVM.SkipReport)
	assert.Equal(t, 10, actualVM.MaxConcurrency)
	tester.AssertThatHasPresets(t, actualVM.JobBase, tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.True(t, actualVM.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualVM.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")
	assert.Equal(t, tester.ImageBootstrap001, actualVM.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma.sh"}, actualVM.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, actualVM.JobBase)
}

func TestKymaIntegrationGKEJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	actualGKE := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-gke-integration", "master")
	require.NotNil(t, actualGKE)
	assert.Equal(t, "kyma-gke-integration", actualGKE.Name)
	assert.Equal(t, "^(resources|installation)", actualGKE.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualGKE, "resources/values.yaml")
	tester.AssertThatJobRunIfChanged(t, *actualGKE, "installation/file.yaml")
	assert.False(t, actualGKE.SkipReport)
	assert.Equal(t, 10, actualGKE.MaxConcurrency)
	tester.AssertThatHasPresets(t, actualGKE.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr, tester.PresetDindEnabled, "preset-sa-gke-kyma-integration", "preset-gc-compute-envs", "preset-docker-push-repository-gke-integration")
	assert.True(t, actualGKE.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualGKE.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualGKE.JobBase.UtilityConfig, "master")
	assert.Equal(t, tester.ImageBootstrapHelm20181121, actualGKE.Spec.Containers[0].Image)
	tester.AssertThatSpecifiesResourceRequests(t, actualGKE.JobBase)
}

func TestKymaIntegrationVMJobPostsubmit(t *testing.T) {
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
	assert.Equal(t, 1, actualVM.MaxConcurrency)
	assert.Equal(t, "", actualVM.RunIfChanged)
	assert.True(t, actualVM.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualVM.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualVM.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualVM.JobBase, tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageBootstrap001, actualVM.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma.sh"}, actualVM.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, actualVM.JobBase)
}

func TestKymaIntegrationGKEJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	kymaPostsubmits, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPostsubmits, 2)

	actualGKE := kymaPostsubmits[1]
	assert.Equal(t, "kyma-gke-integration", actualGKE.Name)
	assert.Equal(t, "", actualGKE.RunIfChanged)
	assert.Equal(t, 1, actualGKE.MaxConcurrency)
	tester.AssertThatHasPresets(t, actualGKE.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildMaster, tester.PresetDindEnabled, "preset-sa-gke-kyma-integration", "preset-gc-compute-envs", "preset-docker-push-repository-gke-integration")
	assert.Equal(t, []string{"master"}, actualGKE.Branches)
	assert.True(t, actualGKE.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualGKE.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualGKE.JobBase.UtilityConfig, "master")
	assert.Equal(t, tester.ImageBootstrapHelm20181121, actualGKE.Spec.Containers[0].Image)
	tester.AssertThatSpecifiesResourceRequests(t, actualGKE.JobBase)
}

func TestKymaIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 5)

	expName := "orphaned-disks-cleaner"
	disksCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, disksCleanerPeriodic)
	assert.Equal(t, expName, disksCleanerPeriodic.Name)
	assert.True(t, disksCleanerPeriodic.Decorate)
	assert.Equal(t, "15 */2 * * *", disksCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, disksCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, disksCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", disksCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, disksCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/disks-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, disksCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, disksCleanerPeriodic.JobBase)

	expName = "orphaned-clusters-cleaner"
	clustersCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, clustersCleanerPeriodic)
	assert.Equal(t, expName, clustersCleanerPeriodic.Name)
	assert.True(t, clustersCleanerPeriodic.Decorate)
	assert.Equal(t, "0 */4 * * *", clustersCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, clustersCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, clustersCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", clustersCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, clustersCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/clusters-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, clustersCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, clustersCleanerPeriodic.JobBase)

	expName = "orphaned-vms-cleaner"
	vmsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, vmsCleanerPeriodic)
	assert.Equal(t, expName, vmsCleanerPeriodic.Name)
	assert.True(t, vmsCleanerPeriodic.Decorate)
	assert.Equal(t, "30 */4 * * *", vmsCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, vmsCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, vmsCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, tester.ImageGolangBuildpackLatest, vmsCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, vmsCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/vms-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, vmsCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, vmsCleanerPeriodic.JobBase)

	expName = "orphaned-loadbalancer-cleaner"
	loadbalancerCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, loadbalancerCleanerPeriodic)
	assert.Equal(t, expName, loadbalancerCleanerPeriodic.Name)
	assert.True(t, loadbalancerCleanerPeriodic.Decorate)
	assert.Equal(t, "30 7 * * 1-5", loadbalancerCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, loadbalancerCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, loadbalancerCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", loadbalancerCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/loadbalancer-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, loadbalancerCleanerPeriodic.JobBase)

}
