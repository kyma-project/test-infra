package compass_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const compassIntegrationTestJobPath = "./../../../../../prow/jobs/incubator/compass/compass-integration.yaml"

func TestCompassIntegrationJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName            string
		expPresets              []preset.Preset
		expRunIfChangedRegex    string
		expRunIfChangedPaths    []string
		expNotRunIfChangedPaths []string
	}{
		"Should contain the compass-integration job": {
			givenJobName: "pre-master-compass-integration",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, preset.BuildPr, "preset-sa-vm-kyma-integration",
			},

			expRunIfChangedRegex: "^(chart|installation)/",
			expRunIfChangedPaths: []string{
				"chart/compass/values.yaml",
				"chart/compass/README.md",
				"installation/cmd/run.sh",
				"installation/resources/installer-cr-with-compass.yaml",
			},
			expNotRunIfChangedPaths: []string{
				"charts/values.yaml",
				"installations/README.md",
				"test/test/test/README.yaml",
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig(compassIntegrationTestJobPath)
			require.NoError(t, err)

			// when
			actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.PresubmitsStatic["kyma-incubator/compass"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
			assert.Equal(t, tc.expRunIfChangedRegex, actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.False(t, actualJob.SkipReport)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualJob.Spec.Containers[0].Image)
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

func TestCompassIntegrationJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []preset.Preset
	}{
		"Should contain the compass-integration job": {
			givenJobName: "post-master-compass-integration",

			expPresets: []preset.Preset{
				preset.GCProjectEnv, preset.KymaGuardBotGithubToken, "preset-sa-vm-kyma-integration",
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig(compassIntegrationTestJobPath)
			require.NoError(t, err)

			// when
			actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.PostsubmitsStatic["kyma-incubator/compass"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, []string{"^master$"}, actualJob.Branches)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.Equal(t, "", actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualJob.Spec.Containers[0].Image)
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
		})
	}
}
