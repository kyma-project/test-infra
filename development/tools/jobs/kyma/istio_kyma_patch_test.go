package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIstioKymaPatchReleases(t *testing.T) {
	// WHEN
	for _, currentRelease := range tester.GetAllKymaReleaseBranches() {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/istio-kyma-patch/istio-kyma-patch.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-components-istio-kyma-patch", currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPresubmit.AlwaysRun)
			assert.Len(t, actualPresubmit.JobBase.Spec.Containers, 1)
			assert.Equal(t, actualPresubmit.JobBase.Spec.Containers[0].Image, tester.ImageBootstrapLatest)
			assert.Len(t, actualPresubmit.JobBase.Spec.Containers[0].Command, 1)
			assert.Equal(t, actualPresubmit.JobBase.Spec.Containers[0].Command[0], "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh")
			assert.Equal(t, actualPresubmit.JobBase.Spec.Containers[0].Args, []string{"/home/prow/go/src/github.com/kyma-project/kyma/components/istio-kyma-patch"})
		})
	}
}

func TestIstioKymaPatchJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/istio-kyma-patch/istio-kyma-patch.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "kyma-components-istio-kyma-patch", "master")
	assert.Len(t, jobConfig.Presubmits, 1)
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "components/istio-kyma-patch/abc")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/components/istio-kyma-patch"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestIstioKymaPatchJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/components/istio-kyma-patch/istio-kyma-patch.yaml")
	// THEN
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], "kyma-components-istio-kyma-patch", "master")
	assert.Len(t, jobConfig.Postsubmits, 1)
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, []string{"master"}, actualPostsubmit.Branches)

	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)
	assert.True(t, actualPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPostsubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^components/istio-kyma-patch/", actualPostsubmit.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/components/istio-kyma-patch"}, actualPostsubmit.Spec.Containers[0].Args)
}
