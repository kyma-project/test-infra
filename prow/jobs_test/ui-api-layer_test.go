package jobs_test

import (
	"github.com/kyma-project/test-infra/prow/jobs_test/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
	"testing"
)

func TestUiApiLayerJobs(t *testing.T) {
	// GIVEN
	jobConfig, err := tester.ReadJobConfig("./../jobs/kyma/components/ui-api-layer/ui-api-layer.yaml")
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	kymaPresubmits, ex := jobConfig.Presubmits["kyma-project/kyma"]

	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 2)

	master := kymaPresubmits[0]
	release := kymaPresubmits[1]

	for _, sut := range []config.Presubmit{master, release} {
		assert.Equal(t, sut.Name, sut.Context)
		assert.True(t, sut.Optional)
		assert.True(t, sut.SkipReport)
		assert.True(t, sut.Decorate)

		tester.AssertThatHasExtraRefTestInfra(t, sut.JobBase.UtilityConfig)
		assert.Equal(t, "test-infra", sut.ExtraRefs[0].Repo)
		assert.Len(t, sut.Spec.Containers, 1)
	}

	assert.Equal(t, "prow/kyma/components/ui-api-layer", master.Name)
	assert.Equal(t, "prow/release/kyma/components/ui-api-layer", release.Name)

	assert.Equal(t, []string{"master"}, master.Branches)
	assert.Equal(t, []string{"^release-\\d+\\.\\d+$"}, release.Branches)

	tester.AssertThatHasPresets(t, master.JobBase, tester.PresetDindEnabled, tester.PresetGcrPush, tester.PresetDockerPushRepo, tester.PresetBuildPr)
	tester.AssertThatHasPresets(t, release.JobBase, tester.PresetDindEnabled, tester.PresetGcrPush, tester.PresetDockerPushRepo, tester.PresetBuildRelease)

}
