package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaileryJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tools/failery/failery.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"})
	assert.Len(t, kymaPresubmits, 1)

	expName := "pre-master-kyma-tools-failery"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(kymaPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr)
	assert.Equal(t, "^tools/failery/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "tools/failery/some_random_file.go"))
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tools/failery"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestFaileryJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tools/failery/failery.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/kyma"})
	assert.Len(t, kymaPost, 1)

	expName := "post-master-kyma-tools-failery"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(kymaPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.BuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, "^tools/failery/", actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tools/failery"}, actualPost.Spec.Containers[0].Args)
}
