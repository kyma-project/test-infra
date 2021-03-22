package testinfra_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProwToolsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-master-test-infra-validate-prow-tools", "master")
	require.NotNil(t, sut)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "development/tools/cmd/configuploader/main.go"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "development/tools/pkg/pkg/file.go"))
	assert.False(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/config.yaml"))

	assert.Equal(t, []string{"^master$", "^main$"}, sut.Branches)
	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangBuildpack1_14, cont.Image)
	assert.Equal(t, []string{"make"}, cont.Command)
	assert.Equal(t, []string{"-C", "development/tools", "validate"}, cont.Args)
}

func TestValidateProwJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-master-test-infra-validate-prow-jobs", "master")
	require.NotNil(t, sut)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/config.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/plugins.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/jobs/random/job.yaml"))
	assert.False(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "development/tools/cmd/configuploader/main.go"))
	assert.False(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "development/tools/pkg/pkg/file.go"))

	assert.Equal(t, []string{"^master$", "^main$"}, sut.Branches)
	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangBuildpack1_14, cont.Image)
	assert.Equal(t, []string{"make"}, cont.Command)
	assert.Equal(t, []string{"-C", "development/tools", "jobs-tests"}, cont.Args)
}

func TestValidateConfigsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-test-infra-validate-configs", "master")
	require.NotNil(t, sut)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/config.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/plugins.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/jobs/random/job.yaml"))

	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageProwToolsCurrent, cont.Image)
	assert.Equal(t,
		[]string{
			"prow/plugins.yaml",
			"prow/config.yaml",
			"prow/jobs",
		},
		cont.Args)
}

func TestValidateScriptsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/validation.yaml")
	// THEN
	require.NoError(t, err)
	testInfraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	sut := tester.FindPresubmitJobByNameAndBranch(testInfraPresubmits, "pre-test-infra-validate-scripts", "master")
	require.NotNil(t, sut)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "development/ala.sh"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*sut, true, "prow/ela.sh"))

	assert.False(t, sut.SkipReport)

	assert.Len(t, sut.Spec.Containers, 1)
	cont := sut.Spec.Containers[0]
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, cont.Image)
	assert.Equal(t, []string{"prow/scripts/validate-scripts.sh"}, cont.Command)
}
