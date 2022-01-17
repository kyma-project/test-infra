package governance_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

func TestDocumentationComponentGovernanceJobPresubmit(t *testing.T) {
	changedFiles := func() ([]string, error) {
		return []string{
			"milv.config.yaml",
			"some_markdown.md",
		}, nil
	}
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/governance.yaml")
	// THEN
	require.NoError(t, err)

	expName := "pre-documentation-component-governance"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/documentation-component"}), expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr, preset.DindEnabled)
	assert.Equal(t, "milv.config.yaml|.md$", actualPresubmit.RunIfChanged)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "milv.config.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "some_markdown.md"))
	shouldRun, err := actualPresubmit.ShouldRun("master", changedFiles, false, true)
	assert.NoError(t, err)
	assert.True(t, shouldRun)

	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--repository", "documentation-component", "--repository-org", "kyma-incubator"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestDocumentationComponentGovernanceJobPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/governance.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	expName := "documentation-component-governance-nightly"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)

	assert.Equal(t, "0 2 * * 1-5", actualPeriodic.Cron)
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.DindEnabled)
	tester.AssertThatHasExtraRepoRefCustom(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra"}, []string{"main"})
	tester.AssertThatHasExtraRef(t, actualPeriodic.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-incubator",
			Repo:      "documentation-component",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-incubator/documentation-component",
		},
	})
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPeriodic.Spec.Containers[0].Command)
	repositoryDirArg := fmt.Sprintf("%s/documentation-component", tester.KymaIncubatorDir)
	assert.Equal(t, []string{"--repository", "documentation-component", "--repository-org", "kyma-incubator", "--repository-dir", repositoryDirArg, "--full-validation", "true"}, actualPeriodic.Spec.Containers[0].Args)
}
