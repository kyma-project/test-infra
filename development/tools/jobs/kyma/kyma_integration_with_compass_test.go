package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaIntegrationWithCompassJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName            string
		expPresets              []tester.Preset
		expRunIfChangedRegex    string
		expRunIfChangedPaths    []string
		expNotRunIfChangedPaths []string
	}{
		"Should contain the kyma-integration job": {
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
		"Should contain the gke-integration job": {
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
		"Should contain the gke-central job": {
			givenJobName: "pre-master-kyma-gke-central-connector",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildPr,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
			expRunIfChangedRegex: "^((resources/core/templates/tests\\S+|resources/application-connector\\S+|installation\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))",
			expRunIfChangedPaths: []string{
				"resources/application-connector/values.yaml",
				"installation/file.yaml",
				"resources/core/templates/tests/test-external-solution.yaml",
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
			// TODO: After the script merge uncomment it!
			// tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
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

func TestKymaIntegrationWithCompassJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []tester.Preset
	}{
		"Should contain the kyma-integration job": {
			givenJobName: "post-master-kyma-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, "preset-sa-vm-kyma-integration",
			},
		},
		"Should contain the gke-integration job": {
			givenJobName: "post-master-kyma-gke-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
		},
		"Should contain the gke-upgrade job": {
			givenJobName: "post-master-kyma-gke-upgrade",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetBuildMaster,
				tester.PresetDindEnabled, tester.PresetKymaGuardBotGithubToken, "preset-sa-gke-kyma-integration",
				"preset-gc-compute-envs", "preset-docker-push-repository-gke-integration",
			},
		},
		"Should contain the gke-central job": {
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
			assert.Equal(t, []string{"^master$"}, actualJob.Branches)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.Equal(t, "", actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			// TODO: After the script merge uncomment it!
			// tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
		})
	}
}
