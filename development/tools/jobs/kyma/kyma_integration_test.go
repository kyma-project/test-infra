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
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-integration", currentRelease), currentRelease)
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

func TestKymaIntegrationGKEJobsReleases(t *testing.T) {
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-gke-integration", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetSaGKEKymaIntegration, tester.PresetGCProjectEnv, tester.PresetBuildRelease, tester.PresetDindEnabled, "preset-sa-gke-kyma-integration", "preset-gc-compute-envs", "preset-docker-push-repository-gke-integration", "preset-kyma-artifacts-bucket")
			assert.False(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.Spec.Containers, 1)
			testContainer := actualPresubmit.Spec.Containers[0]
			assert.Equal(t, tester.ImageBootstrapHelm20181121, testContainer.Image)
			assert.Len(t, testContainer.Command, 1)
			tester.AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
		})
	}
}

func TestKymaIntegrationJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []tester.Preset
		expJobImage  string
	}{
		"Should contains the kyma-integration job": {
			givenJobName: "pre-master-kyma-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration",
			},

			expJobImage: tester.ImageBootstrap001,
		},
		"Should contains the gke-integration job": {
			givenJobName: "pre-master-kyma-gke-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildPr,
				tester.PresetDindEnabled, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
			expJobImage: tester.ImageBootstrapHelm20181121,
		},
		"Should contains the gke-upgrade job": {
			givenJobName: "pre-master-kyma-gke-upgrade",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildPr,
				tester.PresetDindEnabled, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
				"preset-bot-github-token",
			},
			expJobImage: tester.ImageBootstrapHelm20181121,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
			require.NoError(t, err)

			// when
			actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			assert.Equal(t, "^(resources|installation)", actualJob.RunIfChanged)
			tester.AssertThatJobRunIfChanged(t, *actualJob, "resources/values.yaml")
			tester.AssertThatJobRunIfChanged(t, *actualJob, "installation/file.yaml")
			assert.True(t, actualJob.Decorate)
			assert.False(t, actualJob.SkipReport)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tc.expJobImage, actualJob.Spec.Containers[0].Image)
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
		})
	}
}

func TestKymaIntegrationJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []tester.Preset
		expJobImage  string
	}{
		"Should contains the kyma-integration job": {
			givenJobName: "post-master-kyma-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, "preset-sa-vm-kyma-integration",
			},

			expJobImage: tester.ImageBootstrap001,
		},
		"Should contains the gke-integration job": {
			givenJobName: "post-master-kyma-gke-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
			expJobImage: tester.ImageBootstrapHelm20181121,
		},
		"Should contains the gke-upgrade job": {
			givenJobName: "post-master-kyma-gke-upgrade",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
				"preset-bot-github-token",
			},
			expJobImage: tester.ImageBootstrapHelm20181121,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
			require.NoError(t, err)

			// when
			actualJob := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, []string{"master"}, actualJob.Branches)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.Equal(t, "", actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tc.expJobImage, actualJob.Spec.Containers[0].Image)
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
		})
	}
}

func TestKymaIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 8)

	expName := "orphaned-disks-cleaner"
	disksCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, disksCleanerPeriodic)
	assert.Equal(t, expName, disksCleanerPeriodic.Name)
	assert.True(t, disksCleanerPeriodic.Decorate)
	assert.Equal(t, "15 */2 * * 1-5", disksCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, disksCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, disksCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", disksCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, disksCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/disks-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -diskNameRegex='^gke-(gkeint|gke-upgrade).*[-]pvc[-]'"}, disksCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, disksCleanerPeriodic.JobBase)

	expName = "orphaned-clusters-cleaner"
	clustersCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, clustersCleanerPeriodic)
	assert.Equal(t, expName, clustersCleanerPeriodic.Name)
	assert.True(t, clustersCleanerPeriodic.Decorate)
	assert.Equal(t, "0 */4 * * 1-5", clustersCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, clustersCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, clustersCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", clustersCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, clustersCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/clusters-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -clusterNameRegexp='^gke(int|-upgrade)[-](pr|commit|rel)[-].*'"}, clustersCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, clustersCleanerPeriodic.JobBase)

	expName = "orphaned-vms-cleaner"
	vmsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, vmsCleanerPeriodic)
	assert.Equal(t, expName, vmsCleanerPeriodic.Name)
	assert.True(t, vmsCleanerPeriodic.Decorate)
	assert.Equal(t, "30 */4 * * 1-5", vmsCleanerPeriodic.Cron)
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

	expName = "kyma-gke-nightly"
	nightlyPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, nightlyPeriodic)
	assert.Equal(t, expName, nightlyPeriodic.Name)
	assert.True(t, nightlyPeriodic.Decorate)
	assert.Equal(t, "0 4 * * 1-5", nightlyPeriodic.Cron)
	tester.AssertThatHasPresets(t, nightlyPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration, "preset-stability-checker-slack-notifications", "preset-nightly-github-integration")
	tester.AssertThatHasExtraRefs(t, nightlyPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190129-c951cf2", nightlyPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, nightlyPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-long-lasting.sh"}, nightlyPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, nightlyPeriodic.JobBase)
	assert.Len(t, nightlyPeriodic.Spec.Containers[0].Env, 6)
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "nightly")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "TEST_RESULT_WINDOW_TIME", "6h")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "STABILITY_SLACK_CLIENT_CHANNEL_ID", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "developers")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "KYMA_ALERTS_CHANNEL", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnvFromSecret(t, nightlyPeriodic.Spec.Containers[0], "KYMA_ALERTS_SLACK_API_URL", "kyma-alerts-slack-api-url", "secret")

	expName = "kyma-gke-weekly"
	weeklyPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, weeklyPeriodic)
	assert.Equal(t, expName, weeklyPeriodic.Name)
	assert.True(t, weeklyPeriodic.Decorate)
	assert.Equal(t, "0 6 * * 1", weeklyPeriodic.Cron)
	tester.AssertThatHasPresets(t, weeklyPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration, "preset-stability-checker-slack-notifications", "preset-weekly-github-integration")
	tester.AssertThatHasExtraRefs(t, weeklyPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190129-c951cf2", weeklyPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, weeklyPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-long-lasting.sh"}, weeklyPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, weeklyPeriodic.JobBase)
	assert.Len(t, weeklyPeriodic.Spec.Containers[0].Env, 6)
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "weekly")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "TEST_RESULT_WINDOW_TIME", "24h")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "STABILITY_SLACK_CLIENT_CHANNEL_ID", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "developers")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "KYMA_ALERTS_CHANNEL", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnvFromSecret(t, weeklyPeriodic.Spec.Containers[0], "KYMA_ALERTS_SLACK_API_URL", "kyma-alerts-slack-api-url", "secret")

}
