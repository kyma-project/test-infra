package hack_showcase_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlackConnectorJobPresubmit(t *testing.T) {
	//WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/github-slack-connectors/github-slack-connectors.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/github-slack-connectors"})
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(kymaPresubmits, "pre-master-slack-connector", "master")
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, "^slack-connector", actualPresubmit.RunIfChanged)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/github-slack-connectors", actualPresubmit.PathAlias)

	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_12, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/github-slack-connectors/slack-connector", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)

}

func TestSlackConnectorJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/github-slack-connectors/github-slack-connectors.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/github-slack-connectors"})

	actualPost := tester.FindPostsubmitJobByNameAndBranch(kymaPost, "post-master-slack-connector", "master")
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, "^slack-connector", actualPost.RunIfChanged)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/github-slack-connectors", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_12, actualPost.Spec.Containers[0].Image)

	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/github-slack-connectors/slack-connector", "ci-master"}, actualPost.Spec.Containers[0].Args)

}
