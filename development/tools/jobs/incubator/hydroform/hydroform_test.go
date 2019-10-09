package hydroform_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydroformJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/hydroform/hydroform.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	kymaPresubmits, ex := jobConfig.Presubmits["kyma-incubator/hydroform"]
	assert.True(t, ex)
	assert.Len(t, kymaPresubmits, 1)

	actualPresubmit := kymaPresubmits[0]
	expName := "pre-master-hydroform"
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/hydroform", actualPresubmit.PathAlias)
	assert.True(t, actualPresubmit.AlwaysRun)
	assert.Empty(t, actualPresubmit.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.BuildPr)
	assert.Equal(t, tester.ImageGolangBuildpack1_12, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, "GO111MODULE", actualPresubmit.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "on", actualPresubmit.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, "GOPROXY", actualPresubmit.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "https://proxy.golang.org,direct", actualPresubmit.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/hydroform"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestHydroformJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../../prow/jobs/incubator/hydroform/hydroform.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	kymaPost, ex := jobConfig.Postsubmits["kyma-incubator/hydroform"]
	assert.True(t, ex)
	assert.Len(t, kymaPost, 1)

	actualPost := kymaPost[0]
	expName := "post-master-hydroform"
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/hydroform", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.BuildMaster)
	assert.Equal(t, tester.ImageGolangBuildpack1_12, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, "GO111MODULE", actualPost.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "on", actualPost.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, "GOPROXY", actualPost.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "https://proxy.golang.org,direct", actualPost.Spec.Containers[0].Env[0].Value)
	assert.Empty(t, actualPost.RunIfChanged)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/hydroform"}, actualPost.Spec.Containers[0].Args)
}
