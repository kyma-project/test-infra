package compass_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const compassIntegrationTestJobPath = "./../../../../../prow/jobs/incubator/compass/compass-integration.yaml"

func TestCompassIntegrationVMJobsReleases(t *testing.T) {
	// WHEN
	unsupportedReleases := []tester.SupportedRelease{tester.Release12}

	for _, currentRelease := range tester.GetKymaReleaseBranchesBesides(unsupportedReleases) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig(compassIntegrationTestJobPath)
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-incubator/compass"], tester.GetReleaseJobName("compass-integration", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-incubator/compass", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasExtraRefs(t, actualPresubmit.JobBase.UtilityConfig, []string{"cli"})
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, "preset-sa-vm-kyma-integration")
			assert.False(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.Spec.Containers, 1)
			testContainer := actualPresubmit.Spec.Containers[0]
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, testContainer.Image)
			assert.Len(t, testContainer.Command, 1)
			assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-compass.sh", testContainer.Command[0])
			tester.AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
		})
	}
}

func TestCompassIntegrationJobsPresubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName            string
		expPresets              []tester.Preset
		expRunIfChangedRegex    string
		expRunIfChangedPaths    []string
		expNotRunIfChangedPaths []string
	}{
		"Should contain the compass-integration job": {
			givenJobName: "pre-master-compass-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, tester.PresetBuildPr, "preset-sa-vm-kyma-integration",
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
			actualJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-incubator/compass"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
			assert.Equal(t, tc.expRunIfChangedRegex, actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.False(t, actualJob.SkipReport)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatHasExtraRefs(t, actualJob.JobBase.UtilityConfig, []string{"cli"})
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

func TestKymaIntegrationJobsPostsubmit(t *testing.T) {
	tests := map[string]struct {
		givenJobName string
		expPresets   []tester.Preset
	}{
		"Should contain the compass-integration job": {
			givenJobName: "post-master-compass-integration",

			expPresets: []tester.Preset{
				tester.PresetGCProjectEnv, tester.PresetKymaGuardBotGithubToken, "preset-sa-vm-kyma-integration",
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			jobConfig, err := tester.ReadJobConfig(compassIntegrationTestJobPath)
			require.NoError(t, err)

			// when
			actualJob := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/compass"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			// the common expectation
			assert.Equal(t, []string{"^master$"}, actualJob.Branches)
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.Equal(t, "", actualJob.RunIfChanged)
			assert.True(t, actualJob.Decorate)
			assert.Equal(t, "github.com/kyma-incubator/compass", actualJob.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatHasExtraRefs(t, actualJob.JobBase.UtilityConfig, []string{"cli"})
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)

			// the job specific expectation
			assert.Equal(t, tester.ImageKymaClusterInfra20190528, actualJob.Spec.Containers[0].Image)
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
		})
	}
}
