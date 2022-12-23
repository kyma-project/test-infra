package kyma_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

func TestKymaGardenerGCPKyma2ToMainJobPeriodics(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener.yaml")
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	jobName := "kyma-upgrade-gardener-kyma2-to-main-reconciler-main"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, "0 0 6-18/2 ? * 1-5", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 2*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRepoRefCustom(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-upgrade-gardener-kyma2-to-main.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerGCPEventingPresubmit(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener-eventing.yaml")
	require.NoError(t, err)

	presubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"})

	jobName := "pre-main-kyma-gardener-gcp-eventing"
	job := tester.FindPresubmitJobByName(presubmits, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerGCPIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRef(t, job.JobBase.UtilityConfig, []prowapi.Refs{{
		Org:       "kyma-project",
		Repo:      "test-infra",
		BaseRef:   "main",
		PathAlias: "github.com/kyma-project/test-infra",
	}})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener-eventing.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "europe-west4")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "europe-west4-b")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "CREDENTIALS_DIR", "/etc/credentials/kyma-tunas-prow-event-mesh")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
