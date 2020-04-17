package testinfra_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProwPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.Presubmits["kyma-project/test-infra"]

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-master-test-infra-validate-prow", "master")
	require.NotNil(t, sut)

	tester.AssertThatJobRunIfChanged(t, *sut, "development/tools/cmd/configuploader/main.go")
	tester.AssertThatJobRunIfChanged(t, *sut, "development/tools/jobs/console_backend_module_test.go")
	tester.AssertThatJobRunIfChanged(t, *sut, "prow/config.yaml")
	tester.AssertThatJobRunIfChanged(t, *sut, "prow/plugins.yaml")
	tester.AssertThatJobRunIfChanged(t, *sut, "prow/jobs/random/job.yaml")

	assert.Equal(t, []string{"^master$"}, sut.Branches)
	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangBuildpackLatest, cont.Image)
	assert.Equal(t, []string{"make"}, cont.Command)
	assert.Equal(t, []string{"-C", "development/tools", "validate"}, cont.Args)
}

func TestValidateConfigsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.Presubmits["kyma-project/test-infra"]

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-test-infra-validate-configs", "master")
	require.NotNil(t, sut)

	tester.AssertThatJobRunIfChanged(t, *sut, "prow/config.yaml")
	tester.AssertThatJobRunIfChanged(t, *sut, "prow/plugins.yaml")
	tester.AssertThatJobRunIfChanged(t, *sut, "prow/jobs/random/job.yaml")

	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangBuildpack1_13, cont.Image)
	assert.Equal(t,
		[]string{
			"/home/prow/go/src/github.com/kyma-project/test-infra/prow/plugins.yaml",
			"/home/prow/go/src/github.com/kyma-project/test-infra/prow/config.yaml",
			"/home/prow/go/src/github.com/kyma-project/test-infra/prow/jobs",
		},
		cont.Args)
}

func TestValidateScriptsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.Presubmits["kyma-project/test-infra"]

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-test-infra-validate-scripts", "master")
	require.NotNil(t, sut)

	tester.AssertThatJobRunIfChanged(t, *sut, "development/ala.sh")
	tester.AssertThatJobRunIfChanged(t, *sut, "prow/ela.sh")

	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageBootstrapLatest, cont.Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/development/validate-scripts.sh"}, cont.Command)
}
