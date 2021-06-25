package kyma_test

import (
	"testing"
	"time"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

func TestKymaGardenerGCPKymaToKyma2JobPeriodics(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener.yaml")
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	jobName := "kyma-upgrade-gardener-kyma-to-kyma2"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, "0 0 6-18/2 ? * 2-6", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 2*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRepoRefCustom(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-upgrade-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerAzureIntegrationJobPeriodics(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener.yaml")
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	jobName := "kyma-integration-gardener-azure"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, "0 4,7,10,13 * * *", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 4*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRepoRefCustom(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)

	jobName = "kyma-alpha-integration-evaluation-gardener-azure"
	job = tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, "5 * * * *", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 2*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRepoRefCustom(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_ALPHA", "true")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerGCPIntegrationJobPeriodics(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener.yaml")
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	jobName := "kyma-integration-gardener-gcp"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, "00 08 * * *", job.Cron)
	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 4*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerGCPIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRepoRefCustom(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "europe-west4")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "europe-west4-b")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerAWSIntegrationJobPeriodics(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener.yaml")
	require.NoError(t, err)

	periodics := jobConfig.AllPeriodics()

	jobName := "kyma-integration-gardener-aws"
	job := tester.FindPeriodicJobByName(periodics, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.Equal(t, job.DecorationConfig.Timeout.Get(), 4*time.Hour)
	assert.Equal(t, job.DecorationConfig.GracePeriod.Get(), 10*time.Minute)
	assert.Equal(t, "00 14 * * *", job.Cron)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAWSIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRepoRefCustom(t, job.JobBase.UtilityConfig, []string{"test-infra", "kyma"}, []string{"main", "main"})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "eu-west-3")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "eu-west-3a")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}

func TestKymaGardenerAzureIntegrationPresubmit(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-integration-gardener.yaml")
	require.NoError(t, err)

	presubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/kyma"})

	jobName := "pre-main-kyma-gardener-azure-integration"
	job := tester.FindPresubmitJobByName(presubmits, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.True(t, job.Optional)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRef(t, job.JobBase.UtilityConfig, []prowapi.Refs{{
		Org:       "kyma-project",
		Repo:      "test-infra",
		BaseRef:   "main",
		PathAlias: "github.com/kyma-project/test-infra",
	}})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener.sh"}, job.Spec.Containers[0].Command)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)

	jobName = "pre-main-kyma-gardener-azure-alpha-eval"
	job = tester.FindPresubmitJobByName(presubmits, jobName)
	require.NotNil(t, job)
	assert.Equal(t, jobName, job.Name)

	assert.True(t, job.Optional)
	tester.AssertThatHasPresets(t, job.JobBase, preset.GardenerAzureIntegration, preset.KymaCLIStable, preset.ClusterVersion)
	tester.AssertThatHasExtraRef(t, job.JobBase.UtilityConfig, []prowapi.Refs{{
		Org:       "kyma-project",
		Repo:      "test-infra",
		BaseRef:   "main",
		PathAlias: "github.com/kyma-project/test-infra",
	}})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, job.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/kyma-integration-gardener.sh"}, job.Spec.Containers[0].Command)
	assert.Equal(t, "^((resources\\S+|installation\\S+|tools/kyma-installer\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))", job.RunIfChanged)
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_PROJECT_DIR", "/home/prow/go/src/github.com/kyma-project")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "GARDENER_ZONES", "1")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "RS_GROUP", "kyma-gardener-azure")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "REGION", "northeurope")
	tester.AssertThatContainerHasEnv(t, job.Spec.Containers[0], "KYMA_ALPHA", "true")
	tester.AssertThatSpecifiesResourceRequests(t, job.JobBase)
}
