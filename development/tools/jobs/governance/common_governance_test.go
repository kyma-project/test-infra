package governance_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommonGovernanceJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/governance.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 9)
	//presubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/website"})
	//assert.Len(t, presubmits, 1)
}

func TestCommonGovernanceJobPerodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/governance.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 9)

	kyma_presubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"})
	assert.Len(t, kyma_presubmits, 3)
}
