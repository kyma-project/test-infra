package controlplane_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	prowapi "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

const reconcilerIntegrationTestJobPath = "./../../../../prow/jobs/control-plane/control-plane-reconciler-integration.yaml"

func TestReconcilerJobsPresubmitE2E(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(reconcilerIntegrationTestJobPath)
	// THEN
	require.NoError(t, err)

	kymaPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/control-plane"})
	expName := "pre-main-control-plane-reconciler-e2e"
	actualPresubmit := tester.FindPresubmitJobByName(kymaPresubmits, expName)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.False(t, actualPresubmit.Optional)
	assert.False(t, actualPresubmit.AlwaysRun)
	assert.Equal(t, actualPresubmit.RunIfChanged, "^resources/kcp/values.yaml|^resources/kcp/charts/mothership-reconciler/|^resources/kcp/charts/component-reconcilers/")
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "main")
	tester.AssertThatHasExtraRef(t, actualPresubmit.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-project",
			Repo:      "kyma",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-project/kyma",
		},
	})
	tester.AssertThatHasExtraRef(t, actualPresubmit.JobBase.UtilityConfig, []prowapi.Refs{
		{
			Org:       "kyma-incubator",
			Repo:      "reconciler",
			BaseRef:   "main",
			PathAlias: "github.com/kyma-incubator/reconciler",
		},
	})

	assert.Equal(t, tester.ImageKymaIntegrationLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/cluster-integration/reconciler-e2e-gardener.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/reconciler"}, actualPresubmit.Spec.Containers[0].Args)
}
