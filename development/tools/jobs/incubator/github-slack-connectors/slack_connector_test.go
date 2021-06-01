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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(kymaPresubmits, "pre-main-slack-connector", "master")
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, "^slack-connector", actualPresubmit.RunIfChanged)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPresubmit.Spec.Containers[0].Image)
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

	actualPost := tester.FindPostsubmitJobByNameAndBranch(kymaPost, "post-main-slack-connector", "master")
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, "^slack-connector", actualPost.RunIfChanged)
	assert.Equal(t, 10, actualPost.MaxConcurrency)

	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPost.Spec.Containers[0].Image)

	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/github-slack-connectors/slack-connector", "ci-master"}, actualPost.Spec.Containers[0].Args)

}
