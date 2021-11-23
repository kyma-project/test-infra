package reconciler_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

func TestReconcilerJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/reconciler"})
	expName := "pre-main-kyma-incubator-reconciler"
	actualPresubmit := tester.FindPresubmitJobByName(kymaPresubmits, expName)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Empty(t, actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestReconcilerIntegrationJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/reconciler"})
	expName := "pre-main-reconciler-integration-k3d"
	actualPresubmit := tester.FindPresubmitJobByName(kymaPresubmits, expName)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.False(t, actualPresubmit.Optional)
	assert.False(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, actualPresubmit.RunIfChanged, "^((cmd\\S+|configs\\S+|internal\\S+|pkg\\S+)(\\.[^.][^.][^.]+$|\\.[^.][^dD]$|\\.[^mM][^.]$|\\.[^.]$|/[^.]+$))")
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	assert.Equal(t, tester.ImageKymaIntegrationLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/provision-vm-and-start-reconciler-k3d.sh"}, actualPresubmit.Spec.Containers[0].Command)
}

func TestReconcilerJobsPeriodicE2EUpgrade(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	kymaPeriodics := jobConfig.AllPeriodics()
	expName := "periodic-main-kyma-incubator-reconciler-kyma1-kyma2-upgrade"
	actualPeriodic := tester.FindPeriodicJobByName(kymaPeriodics, expName)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.Equal(t, "0 1-22/2 * * 1-5", actualPeriodic.Cron)
	assert.Equal(t, 0, actualPeriodic.JobBase.MaxConcurrency)
	tester.AssertThatHasExtraRefTestInfra(t, actualPeriodic.JobBase.UtilityConfig, "main")
	tester.AssertThatHasExtraRef(t, actualPeriodic.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "kyma",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-project/kyma",
		},
	})
	tester.AssertThatHasExtraRef(t, actualPeriodic.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "kyma",
			BaseRef:   "release-1.24",
			PathAlias: "github.com/kyma-project/kyma-1.24",
		},
	})
	tester.AssertThatHasExtraRef(t, actualPeriodic.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "control-plane",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-project/control-plane",
		},
	})
	assert.Equal(t, tester.ImageKymaIntegrationLatest, actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-e2e-upgrade-gardener.sh"}, actualPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPeriodic.Spec.Containers[0].Args)
}

func TestReconcilerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/reconciler"})
	expName := "post-main-kyma-incubator-reconciler"
	actualPost := tester.FindPostsubmitJobByName(kymaPost, expName)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "main")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush)
	assert.Equal(t, tester.ImageGolangBuildpack1_16, actualPost.Spec.Containers[0].Image)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPost.Spec.Containers[0].Args)
}

func TestReconcilerJobNightlyE2E(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	allPeriodics := jobConfig.AllPeriodics()
	expName := "nightly-main-reconciler-e2e"
	actualNightlyJob := tester.FindPeriodicJobByName(allPeriodics, expName)
	assert.Equal(t, expName, actualNightlyJob.Name)
	assert.Equal(t, "0 1-22/2 * * 1-5", actualNightlyJob.Cron)
	tester.AssertThatHasExtraRef(t, actualNightlyJob.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "control-plane",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-project/control-plane",
		},
	})
	tester.AssertThatHasExtraRef(t, actualNightlyJob.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "kyma",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-project/kyma",
		},
	})
	tester.AssertThatHasExtraRef(t, actualNightlyJob.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "test-infra",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-project/test-infra",
		},
	})
	tester.AssertThatHasExtraRefTestInfra(t, actualNightlyJob.JobBase.UtilityConfig, "main")
	assert.Equal(t, tester.ImageKymaIntegrationLatest, actualNightlyJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-e2e-nightly-gardener.sh"}, actualNightlyJob.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualNightlyJob.Spec.Containers[0].Args)
}
