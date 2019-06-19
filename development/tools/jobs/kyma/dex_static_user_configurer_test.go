package tools_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticUsersGeneratorReleases(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/kyma/components/dex-static-user-configurer/dex-static-user-configurer.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-components-dex-static-user-configurer", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.JobBase.Spec.Containers, 1)
			assert.Equal(t, actualPresubmit.JobBase.Spec.Containers[0].Image, tester.ImageBootstrapLatest)
			assert.Len(t, actualPresubmit.JobBase.Spec.Containers[0].Command, 1)
			assert.Equal(t, actualPresubmit.JobBase.Spec.Containers[0].Command[0], "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh")
			assert.Equal(t, actualPresubmit.JobBase.Spec.Containers[0].Args, []string{"/home/prow/go/src/github.com/kyma-project/kyma/components/dex-static-user-configurer"})
		})
	}
}

func TestStaticUsersGeneratorJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/kyma/components/dex-static-user-configurer/dex-static-user-configurer.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-components-dex-static-user-configurer", "master")
	require.NotNil(t, actualPresubmit)
	assert.Len(t, jobConfig.Presubmits, 1)
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^components/dex-static-user-configurer/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "components/dex-static-user-configurer/some_random_file.go")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/components/dex-static-user-configurer"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestStaticUsersGeneratorJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/kyma/components/dex-static-user-configurer/dex-static-user-configurer.yaml")
	// THEN
	require.NoError(t, err)

	expName := "post-master-kyma-components-dex-static-user-configurer"
	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], expName, "master")
	require.NotNil(t, actualPost)

	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^components/dex-static-user-configurer/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/components/dex-static-user-configurer"}, actualPost.Spec.Containers[0].Args)
}
