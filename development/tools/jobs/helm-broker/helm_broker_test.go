package helm_broker_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelmBrokerJobsPresubmit(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/helm-broker/helm-broker.yaml")
	require.NoError(t, err)

	tests := map[string]struct {
		givenJobName string

		expPresets      []preset.Preset
		expContainerImg string
		expCommand      string
		expArgs         []string
	}{
		"pre-master-helm-broker": {
			givenJobName: "pre-master-helm-broker",

			expPresets: []preset.Preset{
				preset.DindEnabled, preset.GcrPush, preset.DockerPushRepoKyma,
			},
			expContainerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			expCommand:      "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh",
			expArgs:         []string{"/home/prow/go/src/github.com/kyma-project/helm-broker", "ci-pr"},
		},
		"pre-master-helm-broker-chart-test": {
			givenJobName: "pre-master-helm-broker-chart-test",

			expPresets: []preset.Preset{
				preset.DindEnabled, preset.GcrPush, preset.DockerPushRepoKyma, preset.KindVolumesMounts,
			},
			expContainerImg: tester.ImageGolangBuildpack1_14,
			expCommand:      "make",
			expArgs:         []string{"charts-test"},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/helm-broker"}), tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.False(t, actualJob.SkipReport)
			assert.True(t, actualJob.Decorate)
			assert.True(t, actualJob.AlwaysRun)
			assert.Empty(t, actualJob.RunIfChanged)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
			assert.Equal(t, tc.expContainerImg, actualJob.Spec.Containers[0].Image)
			assert.Equal(t, []string{tc.expCommand}, actualJob.Spec.Containers[0].Command)
			assert.Equal(t, tc.expArgs, actualJob.Spec.Containers[0].Args)
		})
	}
}

func TestHelmBrokerJobsPostsubmits(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/helm-broker/helm-broker.yaml")
	require.NoError(t, err)
	assert.Len(t, jobConfig.PostsubmitsStatic, 1)

	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/helm-broker"})
	assert.Len(t, kymaPost, 2)

	for i, tests := range []struct {
		expName         string
		expPresets      []preset.Preset
		expBranches     []string
		expContainerImg string
		expCommand      string
		expArgs         []string
	}{
		{
			expName:         "post-master-helm-broker",
			expBranches:     []string{"^master$", "^main$"},
			expPresets:      []preset.Preset{preset.DindEnabled, preset.GcrPush, preset.DockerPushRepoKyma},
			expContainerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			expCommand:      "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh",
			expArgs:         []string{"/home/prow/go/src/github.com/kyma-project/helm-broker", "ci-master"},
		},
		{
			expName:         "post-release-helm-broker",
			expBranches:     []string{"v\\d+\\.\\d+\\.\\d+$"},
			expPresets:      []preset.Preset{preset.DindEnabled, preset.GcrPush, preset.DockerPushRepoKyma, preset.BotGithubToken, preset.KindVolumesMounts},
			expContainerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			expCommand:      "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh",
			expArgs:         []string{"/home/prow/go/src/github.com/kyma-project/helm-broker", "ci-release"},
		},
	} {
		t.Run(tests.expName, func(t *testing.T) {
			actualPost := kymaPost[i]
			assert.Equal(t, tests.expName, actualPost.Name)
			assert.Equal(t, tests.expBranches, actualPost.Branches)

			assert.Equal(t, 10, actualPost.MaxConcurrency)
			assert.True(t, actualPost.Decorate)
			tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
			tester.AssertThatHasPresets(t, actualPost.JobBase, tests.expPresets...)
			assert.Equal(t, tests.expContainerImg, actualPost.Spec.Containers[0].Image)
			assert.Empty(t, actualPost.RunIfChanged)
			assert.Equal(t, []string{tests.expCommand}, actualPost.Spec.Containers[0].Command)
			assert.Equal(t, tests.expArgs, actualPost.Spec.Containers[0].Args)
		})
	}

}
