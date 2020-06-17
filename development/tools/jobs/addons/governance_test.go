package addons_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"fmt"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddonsGovernanceJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/addons/addons-governance.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	presubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/addons"})
	assert.Len(t, presubmits, 1)

	expName := "pre-master-addons-governance"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(presubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/addons", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr, preset.DindEnabled)
	assert.Equal(t, "milv.config.yaml|.md$", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "milv.config.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "some_markdown.md"))
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--repository", "addons"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestAddonsGovernanceJobPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/addons/addons-governance.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 1)

	expName := "addons-governance-nightly"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "0 1 * * 1-5", actualPeriodic.Cron)
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.DindEnabled)
	tester.AssertThatHasExtraRepoRef(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra", "addons"})
	assert.Equal(t, tester.ImageBootstrapLatest, actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPeriodic.Spec.Containers[0].Command)
	repositoryDirArg := fmt.Sprintf("%s/addons", tester.KymaProjectDir)
	assert.Equal(t, []string{"--repository", "addons", "--repository-dir", repositoryDirArg, "--full-validation", "true"}, actualPeriodic.Spec.Containers[0].Args)
}
