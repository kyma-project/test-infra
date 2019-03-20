package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const e2eTestUpgradeJobPath = "./../../../../prow/jobs/kyma/tests/end-to-end/upgrade/upgrade.yaml"

func TestEnd2EndUpgradeTestsJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(e2eTestUpgradeJobPath)

	// then
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-tests-end-to-end-upgrade", "master")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^tests/end-to-end/upgrade/[^chart]", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/main.go")
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/internal/pkg/pkg.go")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/chart/readme.md")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/chart/upgrade/Chart.yaml")
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/upgrade"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestEnd2EndUpgradeTestsJobPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(e2eTestUpgradeJobPath)
	// then
	require.NoError(t, err)

	actualPresubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], "post-master-kyma-tests-end-to-end-upgrade", "master")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^tests/end-to-end/upgrade/[^chart]", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/main.go")
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/internal/pkg/pkg.go")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/chart/readme.md")
	tester.AssertThatJobDoesNotRunIfChanged(t, *actualPresubmit, "tests/end-to-end/upgrade/chart/upgrade/Chart.yaml")
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/upgrade"}, actualPresubmit.Spec.Containers[0].Args)
}
