package kyma

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaReleasePrImageGuard(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-release-pr-image-guard.yaml")
	// THEN
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/kyma"], "pre-release-pr-image-guard", "release-*")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled)
	assert.Equal(t, tester.ImageBootstrap20181204, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/pr-image-guard.sh"}, actualPresubmit.Spec.Containers[0].Command)
}
