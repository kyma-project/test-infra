package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaTelemetryOperatorPreSubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/telemetry-operator/telemetry-test.yaml")
	// THEN
	require.NoError(t, err)

	preSubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"})
	assert.Len(t, preSubmits, 1)

	jobName := "pre-telemetry-operator-test"
	job := tester.FindPresubmitJobByName(preSubmits, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, tester.ImageGolangBuildpack1_16, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/kyma-telemetry-test.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaTelemetryOperatorPeriodic(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/telemetry-operator/telemetry-test.yaml")
	// THEN
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()
	assert.Len(t, periodics, 1)

	jobName := "telemetry-operator-test"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, "00 07 * * *", job.Cron)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/kyma-telemetry-test.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
