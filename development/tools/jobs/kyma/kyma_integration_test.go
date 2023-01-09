package kyma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

func TestKymaIntegrationJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName            string
		expPresets              []preset.Preset
		expRunIfChangedRegex    string
		expRunIfChangedPaths    []string
		expNotRunIfChangedPaths []string
	}{
		"Should contain the kyma-integration k3d with central Application Connectivity job": {
			givenJobName: "pre-main-kyma-integration-k3d",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, preset.BuildPr, "preset-sa-vm-kyma-integration", "preset-kyma-integration-central-app-connectivity-enabled",
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
		"Should contain the kyma-integration k3d with central Application Connectivity and Compass job": {
			givenJobName: "pre-main-kyma-integration-k3d-central-app-connectivity-compass",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, preset.BuildPr, "preset-sa-vm-kyma-integration", "preset-kyma-integration-central-app-connectivity-enabled", "preset-kyma-integration-compass-dev", "preset-kyma-integration-compass-enabled",
			},
		},
		"Should contain the kyma-integration k3d with telemetry job": {
			givenJobName: "pre-main-kyma-integration-k3d-telemetry",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, preset.BuildPr, "preset-sa-vm-kyma-integration", "preset-kyma-integration-telemetry-enabled",
			},

			expRunIfChangedRegex: "^resources/telemetry/|^installation/resources/crds/telemetry/|^tests/fast-integration/telemetry-test/",
			expRunIfChangedPaths: []string{
				"resources/telemetry/charts/operator/values.yaml",
				"resources/telemetry/charts/fluent-bit/values.yaml",
				"installation/resources/crds/telemetry/logpipelines.crd.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"components/directory-size-exporter/main.go",
				"components/telemetry-operator/main.go",
				"components/webhook-cert-init/main.go",
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

func TestKymaIntegrationJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []preset.Preset
		runIfChanged string
	}{

		"Should contain the kyma-integration-k3d with central Application Connectivity job": {
			givenJobName: "post-main-kyma-integration-k3d",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, "preset-sa-vm-kyma-integration", "preset-kyma-integration-central-app-connectivity-enabled",
			},
		},
		"Should contain the kyma-integration k3d with telemetry job": {
			givenJobName: "post-main-kyma-integration-k3d-telemetry",
			runIfChanged: "^resources/telemetry/|^installation/resources/crds/telemetry/|^tests/fast-integration/telemetry-test/",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, "preset-sa-vm-kyma-integration", "preset-kyma-integration-telemetry-enabled",
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
			assert.Equal(t, tc.runIfChanged, actualJob.RunIfChanged)

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
	assert.Len(t, periodics, 17)

	expName := "kyma-upgrade-k3d-kyma2-to-main"
	kymaUpgradePeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, kymaUpgradePeriodic)
	assert.Equal(t, expName, kymaUpgradePeriodic.Name)

	assert.Equal(t, "0 0 6-18/2 ? * 1-5", kymaUpgradePeriodic.Cron)
	tester.AssertThatHasPresets(t, kymaUpgradePeriodic.JobBase, preset.GCProjectEnv, preset.KymaGuardBotGithubToken, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageKymaIntegrationLatest, kymaUpgradePeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-kyma-upgrade-k3d.sh"}, kymaUpgradePeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string(nil), kymaUpgradePeriodic.Spec.Containers[0].Args)
	tester.AssertThatContainerHasEnv(t, kymaUpgradePeriodic.Spec.Containers[0], "KYMA_PROJECT_DIR", ".")
	tester.AssertThatSpecifiesResourceRequests(t, kymaUpgradePeriodic.JobBase)

	expName = "orphaned-disks-cleaner"
	disksCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, disksCleanerPeriodic)
	assert.Equal(t, expName, disksCleanerPeriodic.Name)

	assert.Equal(t, "30 * * * *", disksCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, disksCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, disksCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, disksCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, disksCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/diskscollector -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -diskNameRegex='^gke-'"}, disksCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, disksCleanerPeriodic.JobBase)

	expName = "orphaned-ips-cleaner"
	addressesCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, addressesCleanerPeriodic)
	assert.Equal(t, expName, addressesCleanerPeriodic.Name)

	assert.Equal(t, "0 1 * * *", addressesCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, addressesCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, addressesCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, addressesCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, addressesCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/ipcleaner -project=${CLOUDSDK_CORE_PROJECT} -dry-run=false -ip-exclude-name-regex='^nightly|nightly-(.*)|weekly|weekly-(.*)|nat-auto-ip'"}, addressesCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, addressesCleanerPeriodic.JobBase)

	expName = "orphaned-assetstore-gcp-bucket-cleaner"
	assetstoreGcpBucketCleaner := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, assetstoreGcpBucketCleaner)
	assert.Equal(t, expName, assetstoreGcpBucketCleaner.Name)

	assert.Equal(t, "00 00 * * *", assetstoreGcpBucketCleaner.Cron)
	tester.AssertThatHasPresets(t, assetstoreGcpBucketCleaner.JobBase, preset.GCProjectEnv, preset.SaProwJobResourceCleaner)
	tester.AssertThatHasExtraRepoRefCustom(t, assetstoreGcpBucketCleaner.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, assetstoreGcpBucketCleaner.Spec.Containers[0].Image)
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
	assert.Equal(t, tester.ImageProwToolsLatest, clustersCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, clustersCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/clusterscollector -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -excluded-clusters=kyma-prow,workload-kyma-prow,nightly,weekly,nightly-20,nightly-21,nightly-22"}, clustersCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, clustersCleanerPeriodic.JobBase)

	expName = "orphaned-vms-cleaner"
	vmsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, vmsCleanerPeriodic)
	assert.Equal(t, expName, vmsCleanerPeriodic.Name)

	assert.Equal(t, "15,45 * * * *", vmsCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, vmsCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, vmsCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, vmsCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, vmsCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/vmscollector -project=${CLOUDSDK_CORE_PROJECT} -vmNameRegexp='gke-nightly-.*|gke-weekly.*|shoot--kyma-prow.*|gke-gke-release-.*' -jobLabelRegexp='kyma-gke-nightly|kyma-gke-nightly-.*|kyma-gke-weekly|kyma-gke-weekly-.*|post-rel.*-kyma-release-candidate' -dryRun=false"}, vmsCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, vmsCleanerPeriodic.JobBase)

	expName = "orphaned-loadbalancer-cleaner"
	loadbalancerCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, loadbalancerCleanerPeriodic)
	assert.Equal(t, expName, loadbalancerCleanerPeriodic.Name)

	assert.Equal(t, "15 * * * *", loadbalancerCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, loadbalancerCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, loadbalancerCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, loadbalancerCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/orphanremover -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, loadbalancerCleanerPeriodic.JobBase)

	expName = "orphaned-dns-cleaner"
	dnsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, dnsCleanerPeriodic)
	assert.Equal(t, expName, dnsCleanerPeriodic.Name)

	assert.Equal(t, "30 * * * *", dnsCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, dnsCleanerPeriodic.JobBase, preset.GCProjectEnv, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, dnsCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, dnsCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, dnsCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/dnscollector -project=${CLOUDSDK_CORE_PROJECT} -dnsZone=${CLOUDSDK_DNS_ZONE_NAME} -ageInHours=2 -regions=${CLOUDSDK_COMPUTE_REGION} -dryRun=false"}, dnsCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, dnsCleanerPeriodic.JobBase)

	expName = "gcr-cleaner-prow-workloads"
	gcrCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, gcrCleanerPeriodic)
	assert.Equal(t, expName, gcrCleanerPeriodic.Name)

	assert.Equal(t, "0 1 * * *", gcrCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, gcrCleanerPeriodic.JobBase, preset.SaGKEKymaIntegration)
	tester.AssertThatHasExtraRepoRefCustom(t, gcrCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	assert.Equal(t, tester.ImageProwToolsLatest, gcrCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, gcrCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "/prow-tools/gcrcleaner --repository=eu.gcr.io/sap-kyma-prow-workloads --age-in-hours=168 --dry-run=false"}, gcrCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, gcrCleanerPeriodic.JobBase)

	expName = "github-issues"
	githubIssuesPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, githubIssuesPeriodic)
	assert.Equal(t, expName, githubIssuesPeriodic.Name)

	assert.Equal(t, "0 6 * * *", githubIssuesPeriodic.Cron)
	assert.Equal(t, tester.ImageProwToolsLatest, githubIssuesPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/github-issues.sh"}, githubIssuesPeriodic.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, githubIssuesPeriodic.JobBase)

	expName = "serverless-function-metrics-generator"
	functionsMetricsPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, functionsMetricsPeriodic)
	assert.Equal(t, expName, functionsMetricsPeriodic.Name)

	assert.Equal(t, "0 0,12 * * *", functionsMetricsPeriodic.Cron)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-serverless-metrics-nightly.sh"},
		functionsMetricsPeriodic.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, functionsMetricsPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "nightly")
	tester.AssertThatContainerHasEnv(t, functionsMetricsPeriodic.Spec.Containers[0], "CLUSTER_PROVIDER", "gcp")
}
