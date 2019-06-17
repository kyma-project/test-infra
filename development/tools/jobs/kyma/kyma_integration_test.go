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
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, "preset-sa-vm-kyma-integration")
			assert.False(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.Spec.Containers, 1)
			testContainer := actualPresubmit.Spec.Containers[0]
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, testContainer.Image)
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
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetSaGKEKymaIntegration, tester.PresetGCProjectEnv, tester.PresetBuildRelease, tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration", "preset-gc-compute-envs", "preset-docker-push-repository-gke-integration", "preset-kyma-artifacts-bucket")
			assert.False(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.Spec.Containers, 1)
			testContainer := actualPresubmit.Spec.Containers[0]
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, testContainer.Image)
			assert.Len(t, testContainer.Command, 1)
			tester.AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
		})
	}
}

func TestKymaIntegrationJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName            string
		expPresets              []tester.Preset
		expRunIfChangedRegex    string
		expRunIfChangedPaths    []string
		expNotRunIfChangedPaths []string
	}{
		"Should contains the kyma-integration job": {
			givenJobName: "pre-master-kyma-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, tester.PresetBuildPr, "preset-sa-vm-kyma-integration",
			},

			expRunIfChangedRegex: "^((resources\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/values.yaml",
				"installation/file.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"installation/README.md",
				"installation/test/test/README.MD",
			},
		},
		"Should contains the gke-integration job": {
			givenJobName: "pre-master-kyma-gke-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildPr,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
			expRunIfChangedRegex: "^((resources\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/values.yaml",
				"installation/file.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"installation/README.md",
				"installation/test/test/README.MD",
			},
		},
		"Should contains the gke-central job": {
			givenJobName: "pre-master-kyma-gke-central-connector",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildPr,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
			expRunIfChangedRegex: "^((resources/application-connector\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/application-connector/values.yaml",
				"installation/file.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"installation/README.md",
				"installation/test/test/README.MD",
				"resources/test/values.yaml",
			},
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
			assert.Equal(t, tc.expRunIfChangedRegex, actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.False(t, actualJob.SkipReport)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
			for _, path := range tc.expRunIfChangedPaths {
				tester.AssertThatJobRunIfChanged(t, *actualJob, path)
			}
			for _, path := range tc.expNotRunIfChangedPaths {
				tester.AssertThatJobDoesNotRunIfChanged(t, *actualJob, path)
			}
		})
	}
}

func TestKymaGKEMinioGatewayJobPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-gke-minio-gateway", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.False(t, actualJob.AlwaysRun)
	assert.False(t, actualJob.Optional)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	assert.Equal(t, `^(resources\/assetstore|installation)`, actualJob.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-minio-gateway.sh"}, actualJob.Spec.Containers[0].Args)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-build-pr",
		"preset-sa-gke-kyma-integration", "preset-gc-compute-envs",
		"preset-gc-project-env", "preset-docker-push-repository-gke-integration", "preset-dind-enabled", "preset-kyma-artifacts-bucket")
}

func TestKymaGKEMinioGatewayMigrationJobPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-gke-minio-gateway-migration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.False(t, actualJob.AlwaysRun)
	assert.True(t, actualJob.Optional) // change to assert.False after verification
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	assert.Equal(t, `^(resources\/assetstore|installation)`, actualJob.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-minio-gateway-migration.sh"}, actualJob.Spec.Containers[0].Args)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-build-pr",
		"preset-sa-gke-kyma-integration", "preset-gc-compute-envs",
		"preset-gc-project-env", "preset-docker-push-repository-gke-integration", "preset-dind-enabled", "preset-kyma-artifacts-bucket")
}

// add test here

func TestKymaGKEXipJobPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-xip-integration", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.False(t, actualJob.AlwaysRun)
	// SHOULD BE UNCOMMENTED: tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-integration-xip.sh"}, actualJob.Spec.Containers[0].Args)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-build-pr",
		"preset-sa-gke-kyma-integration", "preset-gc-compute-envs",
		"preset-gc-project-env", "preset-docker-push-repository-gke-integration", "preset-dind-enabled", "preset-kyma-artifacts-bucket")

}

func TestKymaGKEXipUpgradeJobsPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-xip-gke-upgrade", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.False(t, actualJob.AlwaysRun)
	// SHOULD BE UNCOMMENTED: tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"-c", "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-gke-upgrade-xip.sh"}, actualJob.Spec.Containers[0].Args)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-build-pr",
		"preset-sa-gke-kyma-integration", "preset-gc-compute-envs",
		"preset-gc-project-env", "preset-docker-push-repository-gke-integration", "preset-dind-enabled", "preset-kyma-artifacts-bucket")

}

func TestKymaGKEXipCentralConnectorJobsPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-xip-gke-central-connector", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	// SHOULD BE UNCOMMENTED: tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-central-xip.sh"}, actualJob.Spec.Containers[0].Args)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-build-pr", "preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration", "preset-dind-enabled", "preset-kyma-artifacts-bucket")
}

func TestKymaGKEUpgradeJobsPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-gke-upgrade", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.Equal(t, "^((resources\\S+|installation\\S+|tests/end-to-end/upgrade/chart/upgrade/\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualJob, "resources/values.yaml")
	tester.AssertThatJobRunIfChanged(t, *actualJob, "installation/file.yaml")
	tester.AssertThatJobRunIfChanged(t, *actualJob, "tests/end-to-end/upgrade/chart/upgrade/Chart.yaml")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualJob, "tests/end-to-end/upgrade/chart/upgrade/README.md")
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration")
}

func TestKymaGKECentralConnectorJobsPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-gke-central-connector", "master")
	require.NotNil(t, actualJob)

	// then
	assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
	assert.Equal(t, "^((resources/application-connector\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualJob, "resources/application-connector/values.yaml")
	tester.AssertThatJobRunIfChanged(t, *actualJob, "installation/file.yaml")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualJob, "installation/README.md")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualJob, "installation/test/test/README.MD")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualJob, "resources/test/values.yaml")
	assert.True(t, actualJob.Decorate)
	assert.False(t, actualJob.SkipReport)
	assert.Equal(t, 10, actualJob.MaxConcurrency)
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration")
}

func TestKymaBackupTestJobPresubmit(t *testing.T) {
	// given
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration.yaml")
	require.NoError(t, err)

	// when
	actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-backup-test", "master")
	require.NotNil(t, actualJob)

	// then
	assert.True(t, actualJob.Decorate)
	assert.Equal(t, "^((tests/end-to-end/backup-restore-test/deploy/chart/backup-test/\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", actualJob.RunIfChanged)
	tester.AssertThatHasPresets(t, actualJob.JobBase, tester.PresetKymaBackupRestoreBucket, tester.PresetKymaBackupCredentials, tester.PresetGCProjectEnv, tester.PresetBuildPr,
		tester.PresetSaGKEKymaIntegration, "preset-weekly-github-integration")
	tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190129-c951cf2", actualJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, actualJob.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-end-to-end-test.sh"}, actualJob.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
	assert.Len(t, actualJob.Spec.Containers[0].Env, 3)
	tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "INPUT_CLUSTER_NAME", "e2etest")
	tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "REPO_OWNER_GIT", "kyma-project")
	tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "REPO_NAME_GIT", "kyma")
}

func TestKymaIntegrationJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []tester.Preset
	}{
		"Should contains the kyma-integration job": {
			givenJobName: "post-master-kyma-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, "preset-sa-vm-kyma-integration",
			},
		},
		"Should contains the gke-integration job": {
			givenJobName: "post-master-kyma-gke-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
		},
		"Should contains the gke-upgrade job": {
			givenJobName: "post-master-kyma-gke-upgrade",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
		},
		"Should contains the gke-central job": {
			givenJobName: "post-master-kyma-gke-central-connector",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
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
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
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
	assert.Len(t, periodics, 15)

	expName := "orphaned-disks-cleaner"
	disksCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, disksCleanerPeriodic)
	assert.Equal(t, expName, disksCleanerPeriodic.Name)
	assert.True(t, disksCleanerPeriodic.Decorate)
	assert.Equal(t, "30 * * * *", disksCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, disksCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, disksCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", disksCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, disksCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/disks-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -diskNameRegex='^gke-gkeint|gke-upgrade|weekly|nightly|gke-central|gke-minio'"}, disksCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, disksCleanerPeriodic.JobBase)

	expName = "orphaned-assetstore-gcp-bucket-cleaner"
	assetstoreGcpBucketCleaner := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, assetstoreGcpBucketCleaner)
	assert.Equal(t, expName, assetstoreGcpBucketCleaner.Name)
	assert.True(t, assetstoreGcpBucketCleaner.Decorate)
	assert.Equal(t, "00 00 * * *", assetstoreGcpBucketCleaner.Cron)
	tester.AssertThatHasPresets(t, assetstoreGcpBucketCleaner.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, assetstoreGcpBucketCleaner.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", assetstoreGcpBucketCleaner.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, assetstoreGcpBucketCleaner.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/assetstore-gcp-bucket-cleaner.sh -project=${CLOUDSDK_CORE_PROJECT}"}, assetstoreGcpBucketCleaner.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, assetstoreGcpBucketCleaner.JobBase)

	expName = "orphaned-clusters-cleaner"
	clustersCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, clustersCleanerPeriodic)
	assert.Equal(t, expName, clustersCleanerPeriodic.Name)
	assert.True(t, clustersCleanerPeriodic.Decorate)
	assert.Equal(t, "0 * * * *", clustersCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, clustersCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, clustersCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", clustersCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, clustersCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/clusters-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false -whitelisted-clusters=kyma-prow,workload-kyma-prow,nightly,weekly,service-catalog-crd-periodic"}, clustersCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, clustersCleanerPeriodic.JobBase)

	expName = "orphaned-vms-cleaner"
	vmsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, vmsCleanerPeriodic)
	assert.Equal(t, expName, vmsCleanerPeriodic.Name)
	assert.True(t, vmsCleanerPeriodic.Decorate)
	assert.Equal(t, "0 * * * *", vmsCleanerPeriodic.Cron)
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
	assert.Equal(t, "15 * * * *", loadbalancerCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, loadbalancerCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, loadbalancerCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1", loadbalancerCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/loadbalancer-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, loadbalancerCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, loadbalancerCleanerPeriodic.JobBase)

	expName = "firewall-cleaner"
	firewallCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, firewallCleanerPeriodic)
	assert.Equal(t, expName, firewallCleanerPeriodic.Name)
	assert.True(t, firewallCleanerPeriodic.Decorate)
	assert.Equal(t, "45 */4 * * 1-5", firewallCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, firewallCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, firewallCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181119-afd3fbd", firewallCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, firewallCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/firewall-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dryRun=false"}, firewallCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, firewallCleanerPeriodic.JobBase)

	expName = "orphaned-dns-cleaner"
	dnsCleanerPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, dnsCleanerPeriodic)
	assert.Equal(t, expName, dnsCleanerPeriodic.Name)
	assert.True(t, dnsCleanerPeriodic.Decorate)
	assert.Equal(t, "30 * * * *", dnsCleanerPeriodic.Cron)
	tester.AssertThatHasPresets(t, dnsCleanerPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, dnsCleanerPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, tester.ImageGolangBuildpackLatest, dnsCleanerPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, dnsCleanerPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "development/dns-cleanup.sh -project=${CLOUDSDK_CORE_PROJECT} -dnsZone=${CLOUDSDK_DNS_ZONE_NAME} -ageInHours=2 -regions=${CLOUDSDK_COMPUTE_REGION} -dryRun=false"}, dnsCleanerPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, dnsCleanerPeriodic.JobBase)

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
	tester.AssertThatContainerHasEnv(t, nightlyPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")
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
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")
	tester.AssertThatContainerHasEnv(t, weeklyPeriodic.Spec.Containers[0], "KYMA_ALERTS_CHANNEL", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnvFromSecret(t, weeklyPeriodic.Spec.Containers[0], "KYMA_ALERTS_SLACK_API_URL", "kyma-alerts-slack-api-url", "secret")

	expName = "kyma-aks-nightly"
	nightlyAksPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, nightlyAksPeriodic)
	assert.Equal(t, expName, nightlyAksPeriodic.Name)
	assert.True(t, nightlyAksPeriodic.Decorate)
	assert.Equal(t, "0 4 * * 1-5", nightlyAksPeriodic.Cron)
	tester.AssertThatHasPresets(t, nightlyAksPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration, "preset-stability-checker-slack-notifications", "preset-creds-aks-kyma-integration", "preset-docker-push-repository-gke-integration", "preset-nightly-aks-github-integration")
	tester.AssertThatHasExtraRefs(t, nightlyAksPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190528-8897828", nightlyAksPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, nightlyAksPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-aks-long-lasting.sh"}, nightlyAksPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, nightlyAksPeriodic.JobBase)
	assert.Len(t, nightlyAksPeriodic.Spec.Containers[0].Env, 8)
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "RS_GROUP", "kyma-nightly-aks")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "nightly-aks")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "TEST_RESULT_WINDOW_TIME", "6h")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "STABILITY_SLACK_CLIENT_CHANNEL_ID", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")
	tester.AssertThatContainerHasEnv(t, nightlyAksPeriodic.Spec.Containers[0], "KYMA_ALERTS_CHANNEL", "#c4core-kyma-ci-force")
	tester.AssertThatContainerHasEnvFromSecret(t, nightlyAksPeriodic.Spec.Containers[0], "KYMA_ALERTS_SLACK_API_URL", "kyma-alerts-slack-api-url", "secret")

	expName = "kyma-gke-end-to-end-test-backup-restore"
	backupRestorePeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, backupRestorePeriodic)
	assert.Equal(t, expName, backupRestorePeriodic.Name)
	assert.True(t, backupRestorePeriodic.Decorate)
	assert.Equal(t, "0 5 * * 1-5", backupRestorePeriodic.Cron)
	tester.AssertThatHasPresets(t, backupRestorePeriodic.JobBase, tester.PresetKymaBackupRestoreBucket, tester.PresetKymaBackupCredentials, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration, "preset-weekly-github-integration")
	tester.AssertThatHasExtraRefs(t, backupRestorePeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190129-c951cf2", backupRestorePeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, backupRestorePeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-end-to-end-test.sh"}, backupRestorePeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, backupRestorePeriodic.JobBase)
	assert.Len(t, backupRestorePeriodic.Spec.Containers[0].Env, 3)
	tester.AssertThatContainerHasEnv(t, backupRestorePeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "e2etest")
	tester.AssertThatContainerHasEnv(t, backupRestorePeriodic.Spec.Containers[0], "REPO_OWNER_GIT", "kyma-project")
	tester.AssertThatContainerHasEnv(t, backupRestorePeriodic.Spec.Containers[0], "REPO_NAME_GIT", "kyma")

	expName = "kyma-load-tests-weekly"
	loadTestPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, loadTestPeriodic)
	assert.Equal(t, expName, loadTestPeriodic.Name)
	assert.True(t, loadTestPeriodic.Decorate)
	assert.Equal(t, "0 2 * * 1", loadTestPeriodic.Cron)
	tester.AssertThatHasPresets(t, loadTestPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration, "preset-sap-slack-bot-token")
	tester.AssertThatHasExtraRefs(t, loadTestPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190129-c951cf2", loadTestPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, loadTestPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-load-test.sh"}, loadTestPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, loadTestPeriodic.JobBase)
	assert.Len(t, loadTestPeriodic.Spec.Containers[0].Env, 4)
	tester.AssertThatContainerHasEnv(t, loadTestPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "load-test")
	tester.AssertThatContainerHasEnv(t, loadTestPeriodic.Spec.Containers[0], "LOAD_TEST_SLACK_CLIENT_CHANNEL_ID", "#c4-xf-load-test")
	tester.AssertThatContainerHasEnv(t, loadTestPeriodic.Spec.Containers[0], "LT_REQS_PER_ROUTINE", "1600")
	tester.AssertThatContainerHasEnv(t, loadTestPeriodic.Spec.Containers[0], "LT_TIMEOUT", "30")

	expName = "kyma-components-use-recent-versions"
	verTestPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	assert.Equal(t, expName, verTestPeriodic.Name)
	assert.True(t, verTestPeriodic.Decorate)
	assert.Equal(t, "0 4 * * 1", verTestPeriodic.Cron)
	tester.AssertThatHasPresets(t, verTestPeriodic.JobBase, "preset-sap-slack-bot-token")
	tester.AssertThatHasExtraRefs(t, verTestPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/test-infra/buildpack-golang:v20181204-a6e79be", verTestPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, verTestPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/development/tools/scripts/synchronizer-entrypoint.sh ${KYMA_PROJECT_DIR}/test-infra/development"}, verTestPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, verTestPeriodic.JobBase)
	assert.Len(t, verTestPeriodic.Spec.Containers[0].Env, 3)
	tester.AssertThatContainerHasEnv(t, verTestPeriodic.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	//TODO: change to "#c4core-kyma-ci-force" when the component naming convention will be agreed and synchronizer will follow it
	tester.AssertThatContainerHasEnv(t, verTestPeriodic.Spec.Containers[0], "STABILITY_SLACK_CLIENT_CHANNEL_ID", "#test-alert-channel")
	tester.AssertThatContainerHasEnv(t, verTestPeriodic.Spec.Containers[0], "OUT_OF_DATE_DAYS", "3")

	expName = "kyma-gke-service-catalog-crd-periodic"
	scPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, scPeriodic)
	assert.True(t, scPeriodic.Decorate)
	tester.AssertThatHasPresets(t, scPeriodic.JobBase,
		"preset-kyma-keyring",
		"preset-kyma-encryption-key",
		"preset-stability-checker-slack-notifications",
		"preset-service-catalog-crd-periodic-github-integration",
		"preset-sa-gke-kyma-integration",
		"preset-gc-compute-envs",
		"preset-gc-project-env",
		"preset-docker-push-repository-gke-integration",
		"preset-dind-enabled",
		"preset-kyma-artifacts-bucket",
	)
	tester.AssertThatHasExtraRefTestInfra(t, scPeriodic.JobBase.UtilityConfig, "master")
	tester.AssertThatHasExtraRefs(t, scPeriodic.JobBase.UtilityConfig, []string{"kyma"})
	require.Len(t, scPeriodic.Spec.Containers, 1)
	cont := scPeriodic.Spec.Containers[0]
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-cluster-infra:v20190129-c951cf2", cont.Image)
	assert.Equal(t, []string{"bash"}, cont.Command)
	require.Len(t, cont.Args, 2)
	assert.Equal(t, "-c", cont.Args[0])
	assert.Equal(t, "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-long-lasting.sh", cont.Args[1])
	tester.AssertThatContainerHasEnv(t, cont, "INPUT_CLUSTER_NAME", "service-catalog-crd-periodic")
	tester.AssertThatContainerHasEnv(t, cont, "TEST_RESULT_WINDOW_TIME", "24h")
	tester.AssertThatContainerHasEnv(t, cont, "STABILITY_SLACK_CLIENT_CHANNEL_ID", "#c4core-kyma-gopher-pr")
	tester.AssertThatContainerHasEnv(t, cont, "GITHUB_TEAMS_WITH_KYMA_ADMINS_RIGHTS", "cluster-access")
	tester.AssertThatContainerHasEnv(t, cont, "SERVICE_CATALOG_CRD", "true")
	tester.AssertThatContainerHasEnv(t, cont, "KYMA_ALERTS_CHANNEL", "#c4core-kyma-gopher-pr")
	tester.AssertThatContainerHasEnvFromSecret(t, cont, "KYMA_ALERTS_SLACK_API_URL", "kyma-alerts-slack-api-url", "secret")

}
