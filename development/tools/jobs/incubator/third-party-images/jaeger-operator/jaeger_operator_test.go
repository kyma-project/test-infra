package third_party_images_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const jaegerOperatorJobPath = "./../../../../../../prow/jobs/incubator/third-party-images/jaeger-operator/jaeger-operator.yaml"

func TestJaegerOperatorJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(jaegerOperatorJobPath)

	// then
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-incubator/third-party-images"], "pre-master-tpi-jaeger-operator", "master")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "^jaeger-operator/", actualPresubmit.RunIfChanged)
	assert.Equal(t, "github.com/kyma-incubator/third-party-images", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, tester.ImageBootstrap20181204, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/third-party-images/jaeger-operator"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestJaegerOperatorJobPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(jaegerOperatorJobPath)

	// then
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-incubator/third-party-images"], "post-master-tpi-jaeger-operator", "master")
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)
	assert.True(t, actualPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/third-party-images", actualPostsubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildMaster)
	assert.Equal(t, tester.ImageBootstrap20181204, actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/third-party-images/jaeger-operator"}, actualPostsubmit.Spec.Containers[0].Args)
}
