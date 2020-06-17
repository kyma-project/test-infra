package podpreset_crd_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodpresetCrdJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/podpreset-crd/podpreset-crd.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/podpreset-crd"})
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "pre-master-podpreset-crd"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/jpeeler/podpreset-crd", actualPresubmit.PathAlias)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Empty(t, actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, tester.ImageGolangKubebuilderBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/jpeeler/podpreset-crd"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestPodpresetCrdJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/podpreset-crd/podpreset-crd.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/podpreset-crd"})
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-master-podpreset-crd"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/jpeeler/podpreset-crd", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildMaster)
	assert.Equal(t, tester.ImageGolangKubebuilderBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/jpeeler/podpreset-crd"}, actualPost.Spec.Containers[0].Args)
}
