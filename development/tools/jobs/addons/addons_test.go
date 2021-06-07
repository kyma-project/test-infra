package addons_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddonsJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/addons/addons.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/addons"})
	assert.Len(t, kymaPresubmits, 1)

	expName := "pre-main-kyma-addons"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(kymaPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)

	assert.True(t, actualPresubmit.AlwaysRun)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/addons", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.BotGithubToken)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/addons", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestAddonsJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/addons/addons.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/addons"})
	assert.Len(t, kymaPost, 2)

	expName := "post-main-kyma-addons"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(kymaPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)

	assert.Equal(t, "github.com/kyma-project/addons", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.BotGithubToken)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/addons", "ci-master"}, actualPost.Spec.Containers[0].Args)
}

func TestAddonsReleaseJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/addons/addons.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/addons"})
	assert.Len(t, kymaPost, 2)

	expName := "post-rel-kyma-addons"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(kymaPost, expName, "1.2.3")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, 10, actualPost.MaxConcurrency)

	assert.Equal(t, "github.com/kyma-project/addons", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.BotGithubToken)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/addons", "ci-release"}, actualPost.Spec.Containers[0].Args)
	assert.Equal(t, []string{"\\d+\\.\\d+\\.\\d+$"}, actualPost.Branches)
}
