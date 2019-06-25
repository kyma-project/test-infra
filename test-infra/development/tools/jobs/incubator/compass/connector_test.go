package compass_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const connectorJobPath = "./../../../../../prow/jobs/incubator/compass/components/connector/connector.yaml"

func TestConnectorJobReleases(t *testing.T) {
	// WHEN
	unsupportedReleases := []tester.SupportedRelease{tester.Release10, tester.Release11}

	for _, currentRelease := range tester.GetKymaReleaseBranchesBesides(unsupportedReleases) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig(connectorJobPath)
			// THEN
			require.NoError(t, err)

			actualPre := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-incubator/compass"], tester.GetReleaseJobName("compass-components-connector", currentRelease), currentRelease)
			require.NotNil(t, actualPre)

			assert.False(t, actualPre.SkipReport)
			assert.True(t, actualPre.Decorate)
			assert.Equal(t, "github.com/kyma-incubator/compass", actualPre.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPre.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPre.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPre.AlwaysRun)
			tester.AssertThatExecGolangBuildpack(t, actualPre.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-incubator/compass/components/connector")
		})
	}
}

func TestConnectorJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(connectorJobPath)
	// THEN
	require.NoError(t, err)

	actualPre := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-incubator/compass"], "pre-master-compass-components-connector", "master")
	require.NotNil(t, actualPre)

	assert.Equal(t, 10, actualPre.MaxConcurrency)
	assert.False(t, actualPre.SkipReport)
	assert.True(t, actualPre.Decorate)
	assert.False(t, actualPre.Optional)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualPre.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPre.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPre.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^components/connector/", actualPre.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPre, "components/connector/some_random_file.go")
	tester.AssertThatExecGolangBuildpack(t, actualPre.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-incubator/compass/components/connector")
}

func TestConnectorJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(connectorJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/compass"], "post-master-compass-components-connector", "master")
	require.NotNil(t, actualPost)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^components/connector/", actualPost.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPost, "components/connector/some_random_file.go")
	tester.AssertThatExecGolangBuildpack(t, actualPost.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-incubator/compass/components/connector")
}
