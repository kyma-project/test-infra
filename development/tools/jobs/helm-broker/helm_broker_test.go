package helm_broker_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
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
		expArgs         string
	}{
		"pre-master-helm-broker": {
			givenJobName: "pre-master-helm-broker",

			expPresets: []preset.Preset{
				preset.DindEnabled, preset.GcrPush, preset.BuildPr, preset.DockerPushRepoKyma,
			},
			expContainerImg: tester.ImageGolangKubebuilderBuildpackLatest,
			expCommand:      "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh",
			expArgs:         "/home/prow/go/src/github.com/kyma-project/helm-broker",
		},
		"pre-master-helm-broker-chart-test": {
			givenJobName: "pre-master-helm-broker-chart-test",

			expPresets: []preset.Preset{
				preset.DindEnabled, preset.GcrPush, preset.BuildPr, preset.DockerPushRepoKyma, preset.KindVolumesMounts,
			},
			expContainerImg: tester.ImageBootstrapLatest,
			expCommand:      "make",
			expArgs:         "charts-test",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			actualJob := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/helm-broker"], tc.givenJobName, "master")
			require.NotNil(t, actualJob)

			// then
			assert.Equal(t, 10, actualJob.MaxConcurrency)
			assert.False(t, actualJob.SkipReport)
			assert.True(t, actualJob.Decorate)
			assert.Equal(t, "github.com/kyma-project/helm-broker", actualJob.PathAlias)
			assert.True(t, actualJob.AlwaysRun)
			assert.Empty(t, actualJob.RunIfChanged)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, "master")
			tester.AssertThatHasPresets(t, actualJob.JobBase, tc.expPresets...)
			assert.Equal(t, tc.expContainerImg, actualJob.Spec.Containers[0].Image)
			assert.Equal(t, []string{tc.expCommand}, actualJob.Spec.Containers[0].Command)
			assert.Equal(t, []string{tc.expArgs}, actualJob.Spec.Containers[0].Args)
		})
	}
}

func TestHelmBrokerCrdJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/helm-broker/helm-broker.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-project/helm-broker"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-master-helm-broker"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/helm-broker", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.GcrPush, preset.BuildMaster, preset.DockerPushRepoKyma)
	assert.Equal(t, tester.ImageGolangKubebuilderBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/helm-broker"}, actualPost.Spec.Containers[0].Args)
}
