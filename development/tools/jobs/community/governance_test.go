package community_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"fmt"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGovernanceJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/community/community-governance.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	presubmits, ex := jobConfig.Presubmits["kyma-project/community"]
	assert.True(t, ex)
	assert.Len(t, presubmits, 1)

	expName := "pre-master-community-governance"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(presubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/community", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.BuildPr, preset.DindEnabled)
	assert.Equal(t, "milv.config.yaml|.md$", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "milv.config.yaml")
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "some_markdown.md")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--repository", "community"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestGovernanceJobPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/community/community-governance.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	expName := "community-governance-nightly"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "0 1 * * 1-5", actualPeriodic.Cron)
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.DindEnabled)
	tester.AssertThatHasExtraRefs(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra", "community"})
	assert.Equal(t, tester.ImageBootstrapLatest, actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{tester.GovernanceScriptDir}, actualPeriodic.Spec.Containers[0].Command)
	repositoryDirArg := fmt.Sprintf("%s/community", tester.KymaProjectDir)
	assert.Equal(t, []string{"--repository", "community", "--repository-dir", repositoryDirArg, "--full-validation", "true"}, actualPeriodic.Spec.Containers[0].Args)
}
