package common_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommonJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/common/common.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"}), "pre-master-kyma-common", "master")
	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "common/"))
	assert.Equal(t, "^common/", actualPresubmit.RunIfChanged)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_14, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/common", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestCommonJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/common/common.yaml")
	// THEN
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/kyma"}), "post-master-kyma-common", "master")
	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, []string{"^master$", "^main$"}, actualPostsubmit.Branches)
	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)
	assert.True(t, actualPostsubmit.Decorate)
	tester.AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, preset.GcrPush)
	assert.Equal(t, "^common/", actualPostsubmit.RunIfChanged)
	assert.Equal(t, tester.ImageGolangBuildpack1_14, actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/common", "ci-master"}, actualPostsubmit.Spec.Containers[0].Args)
}
