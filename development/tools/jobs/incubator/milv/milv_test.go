package milv_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilvJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/milv/milv.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/milv"})
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "pre-main-kyma-incubator-milv"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Empty(t, actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.16.4", actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/milv"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestMilvJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/milv/milv.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/milv"})
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-main-kyma-incubator-milv"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^main$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)

	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.16.4", actualPost.Spec.Containers[0].Image)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/milv"}, actualPost.Spec.Containers[0].Args)
}
