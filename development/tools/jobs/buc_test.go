package jobs_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/kube"
	"testing"
)

func TestBucJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../prow/jobs/kyma/components/binding-usage-controller/binding-usage-controller.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	kymaPresubmits, ex := jobConfig.Presubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "prow/kyma/components/binding-usage-controller"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, expName, actualPresubmit.Context)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.True(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.True(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "components/binding-usage-controller/", actualPresubmit.RunIfChanged)
	assert.Equal(t, tester.ImageGolangBuildpack, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, kube.EnvVar{Name: tester.EnvSourcesDir, Value: "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller"}, actualPresubmit.Spec.Containers[0].Env[0])
}

func TestBucJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../prow/jobs/kyma/components/binding-usage-controller/binding-usage-controller.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-project/kyma"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "prow/kyma/components/binding-usage-controller"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpack, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, kube.EnvVar{Name: tester.EnvSourcesDir, Value: "/home/prow/go/src/github.com/kyma-project/kyma/components/binding-usage-controller"}, actualPost.Spec.Containers[0].Env[0])

}
