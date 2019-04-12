package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationProxyReleases(t *testing.T) {
	//Component no longer exists, when we will remove support for 0.7
	// then we can remove both this test and application-proxy.yaml file
	currentRelease := tester.Release07

	// WHEN
	t.Run(currentRelease, func(t *testing.T) {
		jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/application-proxy/application-proxy.yaml")
		// THEN
		require.NoError(t, err)
		actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-components-application-proxy", currentRelease), currentRelease)
		require.NotNil(t, actualPresubmit)
		assert.False(t, actualPresubmit.SkipReport)
		assert.True(t, actualPresubmit.Decorate)
		assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
		tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
		tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
		assert.True(t, actualPresubmit.AlwaysRun)
		tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpackLatest, "/home/prow/go/src/github.com/kyma-project/kyma/components/application-proxy")
	})
}
