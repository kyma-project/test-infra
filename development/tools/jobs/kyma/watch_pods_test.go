package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchPodsReleases(t *testing.T) {
	// When we will remove support for release 0.7
	// then we can remove both this test and watch-pods-deprecated.yaml file
	release := tester.Release07

	t.Run(release, func(t *testing.T) {
		jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tools/watch-pods/watch-pods-deprecated.yaml")
		// THEN
		require.NoError(t, err)
		actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-tools-watch-pods", release), release)
		require.NotNil(t, actualPresubmit)
		assert.False(t, actualPresubmit.SkipReport)
		assert.True(t, actualPresubmit.Decorate)
		assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
		tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, release)
		tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
		assert.True(t, actualPresubmit.AlwaysRun)
		tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/tools/watch-pods")
	})
}
