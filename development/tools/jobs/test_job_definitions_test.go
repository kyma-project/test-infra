package jobs_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJobDefinitionsPresubmitJob(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../prow/jobs/test-infra/test-job-definitions.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.Len(t, testInfraPresubmits, 1)
	sut := testInfraPresubmits[0]

	tester.AssertThatRunIfChanged(t, sut, "prow/jobs/kyma/components/ui-api-layer/ui-api-layer.yaml")
	tester.AssertThatRunIfChanged(t, sut, "development/tools/jobs/ui-api-layer_test.go")

	assert.Equal(t,[]string{"master"}, sut.Branches)


}
