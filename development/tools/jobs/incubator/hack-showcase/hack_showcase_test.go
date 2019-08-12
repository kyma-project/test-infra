package hack_showcase_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHackShowcaseJobPresubmit(t *testing.T) {
	//WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/hack-showcase/hack-showcase.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	kymaPresubmits, ex := jobConfig.Presubmits["kyma-incubator/hack-showcase"]
	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "pre-master-hack-showcase"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/hack-showcase", actualPresubmit.PathAlias)
	assert.Empty(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, "^hack-showcase/", actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/hack-showcase"}, actualPresubmit.Spec.Containers[0].Args)

}

func TestHackShowcaseJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/hack-showcase/hack-showcase.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-incubator/hack-showcase"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-master-hack-showcase"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/hack-showcase", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, "^hack-showcase/", actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/hack-showcase"}, actualPost.Spec.Containers[0].Args)


}
