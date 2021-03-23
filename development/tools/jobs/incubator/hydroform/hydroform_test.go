package hydroform_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydroformJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/hydroform/hydroform.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/hydroform"})
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "pre-master-hydroform"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Empty(t, actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled)
	assert.Equal(t, tester.ImageGolangToolboxLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, "GO111MODULE", actualPresubmit.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "on", actualPresubmit.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/hydroform", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestHydroformJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/hydroform/hydroform.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/hydroform"})
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-master-hydroform"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled)
	assert.Equal(t, tester.ImageGolangToolboxLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, "GO111MODULE", actualPost.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "on", actualPost.Spec.Containers[0].Env[0].Value)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/hydroform", "ci-master"}, actualPost.Spec.Containers[0].Args)
}
