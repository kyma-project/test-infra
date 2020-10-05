package kymacli_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cliCompatibilityJobPath = "./../../../../prow/jobs/cli/cli-compatibility.yaml"

func TestKymaCliCompatibilityPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(cliCompatibilityJobPath)
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 2)

	// Compatibility with previous release
	expName := "kyma-cli-compatibility-1"
	actualPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "00 00 * * 0", actualPeriodic.Cron)
	tester.AssertThatHasExtraRepoRef(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra", "cli"})
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.GCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPeriodic.Spec.Containers[0].Image)
	tester.AssertThatSpecifiesResourceRequests(t, actualPeriodic.JobBase)
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "COMPAT_BACKTRACK", "1")
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "GO111MODULE", "on")
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/compatibility-cli.sh"}, actualPeriodic.Spec.Containers[0].Command)

	// Compatibility 2 releases back
	expName = "kyma-cli-compatibility-2"
	actualPeriodic = tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, actualPeriodic)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.True(t, actualPeriodic.Decorate)
	assert.Equal(t, "00 03 * * 0", actualPeriodic.Cron)
	tester.AssertThatHasExtraRepoRef(t, actualPeriodic.JobBase.UtilityConfig, []string{"test-infra", "cli"})
	tester.AssertThatHasPresets(t, actualPeriodic.JobBase, preset.GCProjectEnv, "preset-sa-vm-kyma-integration")
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, actualPeriodic.Spec.Containers[0].Image)
	tester.AssertThatSpecifiesResourceRequests(t, actualPeriodic.JobBase)
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "COMPAT_BACKTRACK", "2")
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "GO111MODULE", "on")
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/compatibility-cli.sh"}, actualPeriodic.Spec.Containers[0].Command)
}
