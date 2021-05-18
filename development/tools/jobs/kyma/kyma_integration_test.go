package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaIntegrationJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName            string
		expPresets              []preset.Preset
		expRunIfChangedRegex    string
		expRunIfChangedPaths    []string
		expNotRunIfChangedPaths []string
	}{
		"Should contain the kyma-integration job": {
			givenJobName: "pre-main-kyma-integration",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, preset.BuildPr, "preset-sa-vm-kyma-integration",
			},

			expRunIfChangedRegex: "^((resources\\S+|installation\\S+|tools/kyma-installer\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/values.yaml",
				"installation/file.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"installation/README.md",
				"installation/test/test/README.MD",
			},
		},
		"Should contain the kyma-integration K3s job": {
			givenJobName: "pre-main-kyma-integration-k3s",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, preset.BuildPr, "preset-sa-vm-kyma-integration",
			},

			expRunIfChangedRegex: "^((tests/fast-integration\\S+|resources\\S+|installation\\S+|tools/kyma-installer\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/values.yaml",
				"installation/file.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"installation/README.md",
				"installation/test/test/README.MD",
			},
		},
		"Should contain the gke-integration job": {
			givenJobName: "pre-main-kyma-gke-integration",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.BuildPr,
				preset.DindEnabled, preset.KymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
			expRunIfChangedRegex: "^((resources\\S+|installation\\S+|tools/kyma-installer\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/values.yaml",
				"installation/file.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"installation/README.md",
				"installation/test/test/README.MD",
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
			require.NoError(t, err)

			// when
			actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"}), tc.givenJobName, "main")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			assert.Equal(t, tc.expRunIfChangedRegex, actualJob.RunIfChanged)

			assert.False(t, actualJob.SkipReport)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "main")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
			for _, path := range tc.expRunIfChangedPaths {
				assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualJob, true, path))
			}
			for _, path := range tc.expNotRunIfChangedPaths {
				assert.False(t, tester.IfPresubmitShouldRunAgainstChanges(*actualJob, true, path))
			}
		})
	}
}

// add test here
func TestKymaGKEUpgradeJobsPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"}), "pre-main-kyma-gke-upgrade", "main")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.Equal(t, "^((resources\\S+|installation\\S+|tests/end-to-end/upgrade/chart/upgrade/\\S+|tools/kyma-installer\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualJob, true, "resources/values.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualJob, true, "installation/file.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualJob, true, "tests/end-to-end/upgrade/chart/upgrade/Chart.yaml"))
	assert.False(t, tester.IfPresubmitShouldRunAgainstChanges(*actualJob, true, "tests/end-to-end/upgrade/chart/upgrade/README.md"))

	assert.False(t, actualJob.SkipReport)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "main")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaIntegrationLatest, actualJob.Spec.Containers[0].Image)
	tester.AssertThatHasPresets(t, actualJob.JobBase, preset.GCProjectEnv, preset.BuildPr,
		preset.DindEnabled, preset.KymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration")
}

func TestKymaIntegrationJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []preset.Preset
	}{
		"Should contain the kyma-integration job": {
			givenJobName: "post-main-kyma-integration",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, "preset-sa-vm-kyma-integration",
			},
		},
		"Should contain the kyma-integration-k3s job": {
			givenJobName: "post-main-kyma-integration-k3s",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, "preset-sa-vm-kyma-integration",
			},
		},
		"Should contain the gke-integration job": {
			givenJobName: "post-main-kyma-gke-integration",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.BuildMaster,
				preset.DindEnabled, preset.KymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
		},
		"Should contain the gke-upgrade job": {
			givenJobName: "post-main-kyma-gke-upgrade",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.BuildMaster,
				preset.DindEnabled, preset.KymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
			require.NoError(t, err)

			// when
			actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/kyma"}), tc.givenJobName, "main")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, []string{"^master$", "^main$"}, actualJob.Branches)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.Equal(t, "", actualJob.RunIfChanged)

			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "main")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
		})
	}
}

func TestKymaIntegrationJobPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 15)

	expName := "orphaned-disks-cleaner"
	disksCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, disksCleanerPeriodic)
	assert.Equal(t, expName, disksCleanerPeriodic.Name)

	assert.Equal(t, "30 * * * *", disksCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, disksCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, disksCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, disksCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, disksCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/diskscollector -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -diskNameRegex='^gke-gkeint|gke-upgrade|gke-provisioner|gke-backup|weekly|nightly|gke-central|gke-minio|gke-gkecompint|restore|gke-release|gke-gkeext|gke-gke-|gke-gkekcpint|gke-test|gke-cli'"}, disksCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, disksCleanerPeriodic.JobBase)

	expName = "orphaned-az-storage-accounts-cleaner"
	orphanedAZStorageAccountsCleaner := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, orphanedAZStorageAccountsCleaner)
	assert.Equal(t, expName, orphanedAZStorageAccountsCleaner.Name)

	assert.Equal(t, "00 00 * * *", orphanedAZStorageAccountsCleaner.Cron)
	tester.AssertThatHasPresets(t, orphanedAZStorageAccountsCleaner.JobBase, "preset-az-kyma-prow-credentials")
	tester.AssertThatHasExtraRepoRefCustom(t, orphanedAZStorageAccountsCleaner.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, orphanedAZStorageAccountsCleaner.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/minio/azure-cleaner.sh"}, orphanedAZStorageAccountsCleaner.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, orphanedAZStorageAccountsCleaner.JobBase)

	expName = "orphaned-assetstore-gcp-bucket-cleaner"
	assetstoreGcpBucketCleaner := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, assetstoreGcpBucketCleaner)
	assert.Equal(t, expName, assetstoreGcpBucketCleaner.Name)

	assert.Equal(t, "00 00 * * *", assetstoreGcpBucketCleaner.Cron)
	tester.AssertThatHasPresets(t, assetstoreGcpBucketCleaner.JobBase, preset.GCProjectEnv, preset.SaProwJobResourceCleaner)
	tester.AssertThatHasExtraRepoRefCustom(t, assetstoreGcpBucketCleaner.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, assetstoreGcpBucketCleaner.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, assetstoreGcpBucketCleaner.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "prow/scripts/assetstore-gcp-bucket-cleaner.sh -project=${CLOUDSDK_CORE_PROJECT}"}, assetstoreGcpBucketCleaner.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, assetstoreGcpBucketCleaner.JobBase)

	expName = "orphaned-clusters-cleaner"
	clustersCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, clustersCleanerPeriodic)
	assert.Equal(t, expName, clustersCleanerPeriodic.Name)

	assert.Equal(t, "0 * * * *", clustersCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, clustersCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, clustersCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, clustersCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, clustersCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/clusterscollector -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -excluded-clusters=kyma-prow,workload-kyma-prow,nightly,weekly"}, clustersCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, clustersCleanerPeriodic.JobBase)

	expName = "orphaned-vms-cleaner"
	vmsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, vmsCleanerPeriodic)
	assert.Equal(t, expName, vmsCleanerPeriodic.Name)

	assert.Equal(t, "15,45 * * * *", vmsCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, vmsCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, vmsCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, vmsCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, vmsCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/vmscollector -project=${CLOUDSDK_CORE_PROJECT} -vmNameRegexp='.*-integration-test-.*|busola-ui-test-.*' -jobLabelRegexp='.*-integration$|busola-ui-test' -dryRun=false"}, vmsCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, vmsCleanerPeriodic.JobBase)

	expName = "orphaned-loadbalancer-cleaner"
	loadbalancerCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, loadbalancerCleanerPeriodic)
	assert.Equal(t, expName, loadbalancerCleanerPeriodic.Name)

	assert.Equal(t, "15 * * * *", loadbalancerCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, loadbalancerCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, loadbalancerCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, loadbalancerCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/orphanremover -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, loadbalancerCleanerPeriodic.JobBase)

	expName = "firewall-cleaner"
	firewallCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, firewallCleanerPeriodic)
	assert.Equal(t, expName, firewallCleanerPeriodic.Name)

	assert.Equal(t, "45 */4 * * 1-5", firewallCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, firewallCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, firewallCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, firewallCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, firewallCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/firewallcleaner -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, firewallCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, firewallCleanerPeriodic.JobBase)

	expName = "orphaned-dns-cleaner"
	dnsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, dnsCleanerPeriodic)
	assert.Equal(t, expName, dnsCleanerPeriodic.Name)

	assert.Equal(t, "30 * * * *", dnsCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, dnsCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, dnsCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsCurrent, dnsCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, dnsCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/dnscollector -project=${CLOUDSDK_CORE_PROJECT} -dnsZone=${CLOUDSDK_DNS_ZONE_NAME} -ageInHours=2 -regions=${CLOUDSDK_COMPUTE_REGION} -dryRun=false"}, dnsCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, dnsCleanerPeriodic.JobBase)

	expName = "github-stats"
	githubStatsPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, githubStatsPeriodic)
	assert.Equal(t, expName, githubStatsPeriodic.Name)

	assert.Equal(t, "0 6 * * *", githubStatsPeriodic.Cron)
	assert.Equal(t, tester.ImageProwToolsCurrent, githubStatsPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/github-stats.sh"}, githubStatsPeriodic.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, githubStatsPeriodic.JobBase)

	expName = "kyma-gke-nightly"
	nightlyPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, nightlyPeriodic)
	assert.Equal(t, expName, nightlyPeriodic.Name)

	assert.Equal(t, "0 4 * * 1-5", nightlyPeriodic.Cron)
	tester.AssertThatHasPresets(t, nightlyPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration, "preset-stability-checker-slack-notifications", "preset-nightly-github-integration", preset.ClusterVersion, "preset-slack-alerts")
	tester.AssertThatHasExtraRepoRefCustom(t, nightlyPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, nightlyPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, nightlyPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-long-lasting.sh"}, nightlyPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, nightlyPeriodic.JobBase)
	assert.Len(t, nightlyPeriodic.Spec.Containers[0].Env, 8)
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "MACHINE_TYPE", "custom-8-15360")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "PROVISION_REGIONAL_CLUSTER", "true")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "NODES_PER_ZONE", "1")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG", "0.6.4")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "nightly")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "TEST_RESULT_WINDOW_TIME", "6h")
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")

	expName = "kyma-gke-weekly"
	weeklyPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, weeklyPeriodic)
	assert.Equal(t, expName, weeklyPeriodic.Name)

	assert.Equal(t, "0 6 * * 1", weeklyPeriodic.Cron)
	tester.AssertThatHasPresets(t, weeklyPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration, "preset-stability-checker-slack-notifications", "preset-weekly-github-integration", preset.ClusterVersion, "preset-slack-alerts")
	tester.AssertThatHasExtraRepoRefCustom(t, weeklyPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, weeklyPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, weeklyPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-long-lasting.sh"}, weeklyPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, weeklyPeriodic.JobBase)
	assert.Len(t, weeklyPeriodic.Spec.Containers[0].Env, 8)
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "MACHINE_TYPE", "custom-12-15360")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "PROVISION_REGIONAL_CLUSTER", "true")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "NODES_PER_ZONE", "1")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "STACKDRIVER_COLLECTOR_SIDECAR_IMAGE_TAG", "0.6.4")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "weekly")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "TEST_RESULT_WINDOW_TIME", "24h")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")

	expName = "kyma-aks-nightly"
	nightlyAksPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, nightlyAksPeriodic)
	assert.Equal(t, expName, nightlyAksPeriodic.Name)

	assert.Equal(t, "0 4 * * 1-5", nightlyAksPeriodic.Cron)
	tester.AssertThatHasPresets(t, nightlyAksPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration, preset.StabilityCheckerSlack, "preset-az-kyma-prow-credentials", "preset-docker-push-repository-gke-integration", "preset-nightly-aks-github-integration", preset.ClusterVersion, "preset-slack-alerts")
	tester.AssertThatHasExtraRepoRefCustom(t, nightlyAksPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, nightlyAksPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, nightlyAksPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-aks-long-lasting.sh"}, nightlyAksPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, nightlyAksPeriodic.JobBase)
	assert.Len(t, nightlyAksPeriodic.Spec.Containers[0].Env, 5)
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "RS_GROUP", "kyma-nightly-aks")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "nightly-aks")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "TEST_RESULT_WINDOW_TIME", "6h")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")

}
