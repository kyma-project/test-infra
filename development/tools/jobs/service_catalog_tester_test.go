package jobs_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/kube"
)

func TestServiceCatalogTesterJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../prow/jobs/service-catalog-tester/service-catalog-tester.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	kymaPresubmits, ex := jobConfig.Presubmits["kyma-incubator/service-catalog-tester"]
	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "prow/service-catalog-tester"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.True(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/service-catalog-tester", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^service-catalog-tester/", actualPresubmit.RunIfChanged)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, kube.EnvVar{Name: tester.EnvSourcesDir, Value: "/home/prow/go/src/github.com/kyma-incubator/service-catalog-tester"}, actualPresubmit.Spec.Containers[0].Env[0])
}

func TestServiceCatalogTesterJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../prow/jobs/service-catalog-tester/service-catalog-tester.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-incubator/service-catalog-tester"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "prow/service-catalog-tester"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/service-catalog-tester", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpackLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, kube.EnvVar{Name: tester.EnvSourcesDir, Value: "/home/prow/go/src/github.com/kyma-incubator/service-catalog-tester"}, actualPost.Spec.Containers[0].Env[0])
}
