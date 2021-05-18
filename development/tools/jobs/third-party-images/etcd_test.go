package third_party_images_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const etcdJobPath = "../../../../prow/jobs/third-party-images/third-party-images.yaml"

func TestEtcdJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(etcdJobPath)

	// then
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/third-party-images"}), "pre-main-tpi-etcd", "main")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "^etcd/", actualPresubmit.RunIfChanged)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoThirdPartyImages, preset.GcrPush)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/third-party-images/etcd", "ci-pr"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestEtcdJobPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(etcdJobPath)

	// then
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/third-party-images"}), "post-main-tpi-etcd", "main")
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)
	assert.True(t, actualPostsubmit.Decorate)
	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoThirdPartyImages, preset.GcrPush)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/third-party-images/etcd", "ci-main"}, actualPostsubmit.Spec.Containers[0].Args)
}
