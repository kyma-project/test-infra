package varkes_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
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
	varkesPresubmits, ex := jobConfig.PresubmitsStatic["kyma-incubator/varkes"]
	assert.True(t, ex)
	assert.Len(t, varkesPresubmits, 1)

	masterPresubmit := tester.FindPresubmitJobByNameAndBranch(varkesPresubmits, jobName, "master")
	expName := jobName
	assert.Equal(t, expName, masterPresubmit.Name)
	assert.Equal(t, []string{"^master$", "release"}, masterPresubmit.Branches)
	assert.Equal(t, 10, masterPresubmit.MaxConcurrency)
	assert.False(t, masterPresubmit.SkipReport)
	assert.True(t, masterPresubmit.Decorate)
	assert.True(t, masterPresubmit.AlwaysRun)
	assert.Equal(t, "github.com/kyma-incubator/varkes", masterPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, masterPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, masterPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, tester.ImageNode10Buildpack, masterPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, masterPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/varkes/"}, masterPresubmit.Spec.Containers[0].Args)
}

func TestVarkesJobMasterPostsubmit(t *testing.T) {
	// WHEN
	const jobName = "post-master-varkes"
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/varkes/varkes.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	varkesPostsubmits, ex := jobConfig.PostsubmitsStatic["kyma-incubator/varkes"]
	assert.True(t, ex)
	assert.Len(t, varkesPostsubmits, 2)

	masterPostsubmit := tester.FindPostsubmitJobByNameAndBranch(varkesPostsubmits, jobName, "master")
	expName := jobName
	assert.Equal(t, expName, masterPostsubmit.Name)
	assert.Equal(t, []string{"^master$"}, masterPostsubmit.Branches)
	assert.Equal(t, 10, masterPostsubmit.MaxConcurrency)
	assert.True(t, masterPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/varkes", masterPostsubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, masterPostsubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, masterPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildMaster)
	assert.Equal(t, tester.ImageNode10Buildpack, masterPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, masterPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/varkes/"}, masterPostsubmit.Spec.Containers[0].Args)
}

func TestVarkesJobReleasePostsubmit(t *testing.T) {
	// WHEN
	const jobName = "post-release-varkes"
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/varkes/varkes.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	varkesPostsubmits, ex := jobConfig.PostsubmitsStatic["kyma-incubator/varkes"]
	assert.True(t, ex)
	assert.Len(t, varkesPostsubmits, 2)

	releasePostsubmit := tester.FindPostsubmitJobByNameAndBranch(varkesPostsubmits, jobName, "release")
	expName := jobName
	assert.Equal(t, expName, releasePostsubmit.Name)
	assert.Equal(t, []string{"release"}, releasePostsubmit.Branches)
	assert.Equal(t, 10, releasePostsubmit.MaxConcurrency)
	assert.True(t, releasePostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/varkes", releasePostsubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, releasePostsubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, releasePostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, tester.ImageNode10Buildpack, releasePostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, releasePostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/varkes/"}, releasePostsubmit.Spec.Containers[0].Args)
}
