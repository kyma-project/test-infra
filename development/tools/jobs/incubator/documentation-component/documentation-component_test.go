package documentation_component_test

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentationComponentJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/documentation-component/documentation-component.yaml")
	// THEN
	require.NoError(t, err)
	expName := "pre-documentation-component"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/documentation-component"}), expName, "master")

	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-incubator/documentation-component", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.BuildPr)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "add-ons/some_random_file.js"))

	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/buildpack-node:v20190724-d41df0f", actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/documentation-component"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestGovernanceJobPresubmit(t *testing.T) {
	changedFiles := func() ([]string, error) {
		return []string{
			"milv.config.yaml",
			"some_markdown.md",
		}, nil
	}
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/documentation-component/documentation-component.yaml")
	// THEN
	require.NoError(t, err)

	expName := "pre-documentation-component-governance"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/documentation-component"}), expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/documentation-component", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr, preset.DindEnabled)
	assert.Equal(t, "milv.config.yaml|.md$", actualPresubmit.RunIfChanged)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "milv.config.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "some_markdown.md"))
	shouldRun, err := actualPresubmit.ShouldRun("master", changedFiles, false, true)
	assert.NoError(t, err)
	assert.True(t, shouldRun)

	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--repository", "documentation-component", "--repository-org", "kyma-incubator"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestGovernanceJobPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/documentation-component/documentation-component.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	expName := "documentation-component-governance-nightly"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "0 1 * * 1-5", actualPeriodic.Cron)
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.DindEnabled)
	tester.AssertThatHasExtraRefs(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra", "documentation-component"})
	assert.Equal(t, tester.ImageBootstrapLatest, actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPeriodic.Spec.Containers[0].Command)
	repositoryDirArg := fmt.Sprintf("%s/documentation-component", tester.KymaIncubatorDir)
	assert.Equal(t, []string{"--repository", "documentation-component", "--repository-org", "kyma-incubator", "--repository-dir", repositoryDirArg, "--full-validation", "true"}, actualPeriodic.Spec.Containers[0].Args)
}
