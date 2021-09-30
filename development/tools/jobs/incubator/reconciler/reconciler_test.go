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

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/reconciler"})
	assert.Len(t, kymaPresubmits, 3)

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

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/reconciler"})
	assert.Len(t, kymaPresubmits, 3)

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

func TestReconcilerJobsPresubmitE2E(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/reconciler"})
	assert.Len(t, kymaPresubmits, 3)

	expName := "pre-main-kyma-incubator-reconciler-e2e"
	actualPresubmit := tester.FindPresubmitJobByName(kymaPresubmits, expName)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Optional)
	assert.False(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, actualPresubmit.RunIfChanged, "^resources")
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")

	// @TODO: For testing using pr image name. Update to tester.xxx
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-integration:v20210902-035ae0cc-k8s1.18", actualPresubmit.Spec.Containers[0].Image)
	//assert.Equal(t, tester.ImageKymaIntegrationLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-e2e-gardener.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestReconcilerJobsPeriodicE2EUpgrade(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Periodics, 3)
	kymaPeriodics := jobConfig.AllPeriodics()
	assert.Len(t, kymaPeriodics, 3)

	expName := "periodic-main-kyma-incubator-reconciler-kyma1-kyma2-upgrade"
	actualPeriodic := tester.FindPeriodicJobByName(kymaPeriodics, expName)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.Equal(t, "30 * * * *", actualPeriodic.Cron)
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
			Org:       "kyma-incubator",
			Repo:      "reconciler",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-incubator/reconciler",
		},
	})

	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-integration:v20210902-035ae0cc-k8s1.18", actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-e2e-upgrade-gardener.sh"}, actualPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPeriodic.Spec.Containers[0].Args)
}

func TestReconcilerJobsNightlyMain(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Periodics, 3)
	reconcilerPeriodics := jobConfig.AllPeriodics()
	assert.Len(t, reconcilerPeriodics, 3)

	expName := "nightly-main-reconciler"
	actualPeriodic := tester.FindPeriodicJobByName(reconcilerPeriodics, expName)
	assert.Equal(t, expName, actualPeriodic.Name)
	assert.Equal(t, "0 0 * * *", actualPeriodic.Cron)
	assert.Equal(t, 0, actualPeriodic.JobBase.MaxConcurrency)

	tester.AssertThatHasExtraRefTestInfra(t, actualPeriodic.JobBase.UtilityConfig, "main")
	tester.AssertThatHasExtraRef(t, actualPeriodic.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-incubator",
			Repo:      "reconciler",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-incubator/reconciler",
		},
	})

	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/kyma-integration:v20210902-035ae0cc-k8s1.18", actualPeriodic.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-gardener-long-lasting.sh"}, actualPeriodic.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPeriodic.Spec.Containers[0].Args)
	tester.AssertThatContainerHasEnv(t, actualPeriodic.Spec.Containers[0], "INPUT_CLUSTER_NAME", "rec-nightly")
}

func TestReconcilerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	kymaPost := jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/reconciler"})
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-main-kyma-incubator-reconciler"
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
	assert.Len(t, allPeriodics, 3)

	expName := "nightly-main-reconciler-e2e"
	actualNightlyJob := tester.FindPeriodicJobByName(allPeriodics, expName)
	assert.Equal(t, expName, actualNightlyJob.Name)
	assert.Equal(t, "0 1-22/2 * * *", actualNightlyJob.Cron)
	tester.AssertThatHasExtraRef(t, actualNightlyJob.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-incubator",
			Repo:      "reconciler",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-incubator/reconciler",
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
	// TODO: use a non-pr tagged image
	assert.Equal(t, "eu.gcr.io/kyma-project/test-infra/pr/kyma-integration:v20210902-035ae0cc-k8s1.18", actualNightlyJob.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-e2e-nightly-gardener.sh"}, actualNightlyJob.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualNightlyJob.Spec.Containers[0].Args)
	tester.AssertThatContainerHasEnv(t, actualNightlyJob.Spec.Containers[0], "INPUT_CLUSTER_NAME", "rec-nightly")
}
