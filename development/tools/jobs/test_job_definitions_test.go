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

	tester.AssertThatJobRunIfChanged(t, sut, "prow/jobs/kyma/components/ui-api-layer/ui-api-layer.yaml")
	tester.AssertThatJobRunIfChanged(t, sut, "development/tools/jobs/ui-api-layer_test.go")

	assert.Equal(t, []string{"master"}, sut.Branches)
	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/development/tools/", cont.WorkingDir)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, cont.Image)
	assert.Equal(t, []string{"make"}, cont.Command)
	assert.Equal(t, []string{"jobs-tests"}, cont.Args)

}
