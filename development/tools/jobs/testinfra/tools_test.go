package testinfra_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/development/tools/tools.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.Len(t, testInfraPresubmits, 1)
	sut := testInfraPresubmits[0]

	tester.AssertThatJobRunIfChanged(t, sut, "development/tools/cmd/configuploader/main.go")
	tester.AssertThatJobRunIfChanged(t, sut, "development/tools/jobs/ui-api-layer_test.go")

	assert.Equal(t, []string{"master"}, sut.Branches)
	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangBuildpackLatest, cont.Image)
	assert.Equal(t, []string{"make"}, cont.Command)
	assert.Equal(t, []string{"-C", "development/tools", "validate"}, cont.Args)

}
