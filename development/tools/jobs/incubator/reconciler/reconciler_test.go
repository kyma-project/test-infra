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
	assert.Len(t, kymaPresubmits, 2)

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

func TestReconcilerJobsPresubmitE2E(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/reconciler/reconciler.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-incubator/reconciler"})
	assert.Len(t, kymaPresubmits, 2)

	expName := "pre-main-kyma-incubator-reconciler-e2e"
	actualPresubmit := tester.FindPresubmitJobByName(kymaPresubmits, expName)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

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

	assert.Len(t, jobConfig.Periodics, 1)
	kymaPeriodics := jobConfig.AllPeriodics()
	assert.Len(t, kymaPeriodics, 1)

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
