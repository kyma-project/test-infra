package console_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoreJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/console/console-integration.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	kymaPresubmits, ex := jobConfig.Presubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 1)

	expName := "console-integration"
	actualPresubmit := tester.FindPresubmitJobByName(kymaPresubmits, expName)
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.True(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig)
	assert.Equal(t, []string{"kyma-project"}, actualPresubmit.Extra_refs[1].Org)
	assert.Equal(t, []string{"console"}, actualPresubmit.Extra_refs[1].Repo)
	assert.Equal(t, []string{"master"}, actualPresubmit.Extra_refs[1].Base_ref)
	assert.Equal(t, []string{"github.com/kyma-project/console"}, actualPresubmit.Extra_refs[1].Path_alias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^components/|^resources/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "components/some_random_file.js")
	assert.Equal(t, tester.ImageNodeBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/console/tests"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestCoreJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/console/core/console-integration.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	expName := "console-integration"
	actualPost := tester.FindPostsubmitJobByName(kymaPost, expName)
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig)
	assert.Equal(t, []string{"kyma-project"}, actualPost.Extra_refs[1].Org)
	assert.Equal(t, []string{"console"}, actualPost.Extra_refs[1].Repo)
	assert.Equal(t, []string{"master"}, actualPost.Extra_refs[1].Base_ref)
	assert.Equal(t, []string{"github.com/kyma-project/console"}, actualPost.Extra_refs[1].Path_alias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, tester.ImageNodeBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/console/tests"}, actualPost.Spec.Containers[0].Args)
}
