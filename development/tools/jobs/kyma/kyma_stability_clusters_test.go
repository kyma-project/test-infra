package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaStabilityClustersPeriodics(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-stability-clusters.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.Periodics
	assert.Len(t, periodics, 1)

	expName = "kyma-gke-nightly"
	nightlyPeriodic := tester.FindPeriodicJobByName(periodics, expName)
	require.NotNil(t, nightlyPeriodic)
	assert.Equal(t, expName, nightlyPeriodic.Name)
	assert.True(t, nightlyPeriodic.Decorate)
	assert.Equal(t, "0 8-16/3 * * 1-5", nightlyPeriodic.Cron)
	tester.AssertThatHasPresets(t, nightlyPeriodic.JobBase, tester.PresetGCProjectEnv, tester.PresetSaGKEKymaIntegration)
	tester.AssertThatHasExtraRefs(t, nightlyPeriodic.JobBase.UtilityConfig, []string{"test-infra", "kyma"})
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/test-infra/bootstrap-helm:v20181121-f2f12bc", nightlyPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"bash"}, nightlyPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-nightly.sh"}, nightlyPeriodic.Spec.Containers[0].Args)
	tester.AssertThatSpecifiesResourceRequests(t, nightlyPeriodic.JobBase)

}
