package kyma_test

import (
	"testing"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestWebsiteJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/docs/website.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-project/website"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	expName := "post-master-kyma-website"
	actualPost := tester.FindPostsubmitJobByName(kymaPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/website", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetBuildMaster, tester.PresetWebsiteBotGithubIdentity, tester.PresetWebsiteBotGithubSSH, tester.PresetWebsiteBotGithubToken)
	assert.Equal(t, tester.ImageNodeBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/website"}, actualPost.Spec.Containers[0].Args)
}
