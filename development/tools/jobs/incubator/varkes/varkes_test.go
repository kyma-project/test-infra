package varkes_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVarkesJobPresubmit(t *testing.T) {
	// WHEN
	const jobName = "pre-varkes"
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/varkes/varkes.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	varkesPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/varkes"})
	assert.Len(t, varkesPresubmits, 1)

	masterPresubmit := tester.FindPresubmitJobByNameAndBranch(varkesPresubmits, jobName, "master")
	expName := jobName
	assert.Equal(t, expName, masterPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, masterPresubmit.Branches)
	assert.Equal(t, 10, masterPresubmit.MaxConcurrency)
	assert.False(t, masterPresubmit.SkipReport)

	assert.True(t, masterPresubmit.AlwaysRun)
	tester.AssertThatHasExtraRefTestInfra(t, masterPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, masterPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageNodeBuildpackLatest, masterPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, masterPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/varkes/", "ci-pr"}, masterPresubmit.Spec.Containers[0].Args)
}

func TestVarkesJobMasterPostsubmit(t *testing.T) {
	// WHEN
	const jobName = "post-main-varkes"
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/varkes/varkes.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	varkesPostsubmits := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/varkes"})
	assert.Len(t, varkesPostsubmits, 2)

	masterPostsubmit := tester.FindPostsubmitJobByNameAndBranch(varkesPostsubmits, jobName, "master")
	expName := jobName
	assert.Equal(t, expName, masterPostsubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, masterPostsubmit.Branches)
	assert.Equal(t, 10, masterPostsubmit.MaxConcurrency)

	tester.AssertThatHasExtraRefTestInfra(t, masterPostsubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, masterPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageNodeBuildpackLatest, masterPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, masterPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/varkes/", "ci-master"}, masterPostsubmit.Spec.Containers[0].Args)
}

func TestVarkesJobReleasePostsubmit(t *testing.T) {
	// WHEN
	const jobName = "post-release-varkes"
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/varkes/varkes.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	varkesPostsubmits := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/varkes"})
	assert.Len(t, varkesPostsubmits, 2)

	releasePostsubmit := tester.FindPostsubmitJobByNameAndBranch(varkesPostsubmits, jobName, "release")
	expName := jobName
	assert.Equal(t, expName, releasePostsubmit.Name)
	assert.Equal(t, []string{"^\\d+\\.\\d+\\.\\d+$"}, releasePostsubmit.Branches)
	assert.Equal(t, 10, releasePostsubmit.MaxConcurrency)

	tester.AssertThatHasExtraRefTestInfra(t, releasePostsubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, releasePostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageNodeBuildpackLatest, releasePostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, releasePostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/varkes/", "ci-release"}, releasePostsubmit.Spec.Containers[0].Args)
}
