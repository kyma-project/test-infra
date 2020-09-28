package controlplane_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKCPPresubmitDevelopmentArtifacts(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/control-plane/kcp-development-artifacts.yaml")
	// THEN
	require.NoError(t, err)

	job := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-project/control-plane"}), "pre-master-kcp-development-artifacts", "master")
	require.NotNil(t, job)

	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*job, true, "resources/provisioner/values.yaml"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*job, true, "installation/scripts/concat-yamls.sh"))
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*job, true, "tools/kcp-installer/kcp.Dockerfile"))
	assert.False(t, job.SkipReport)
	assert.False(t, job.AlwaysRun)
	assert.True(t, job.Optional)
	tester.AssertThatHasExtraRefTestInfra(t, job.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, "preset-kyma-development-artifacts-bucket", preset.GcrPush)
	require.Len(t, job.Spec.Containers, 1)
	cont := job.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangToolboxLatest, cont.Image)
	require.Len(t, cont.Command, 1)
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-kcp-development-artifacts.sh", cont.Command[0])
	require.Len(t, job.Spec.Volumes, 1)
	assert.Equal(t, "sa-kyma-artifacts", job.Spec.Volumes[0].Name)
	require.NotNil(t, job.Spec.Volumes[0].Secret)
	assert.Equal(t, "sa-kyma-artifacts", job.Spec.Volumes[0].Secret.SecretName)
	require.Len(t, cont.VolumeMounts, 1)
	assert.Equal(t, "sa-kyma-artifacts", cont.VolumeMounts[0].Name)
	assert.Equal(t, "/etc/credentials/sa-kyma-artifacts", cont.VolumeMounts[0].MountPath)
}

func TestKCPPostsubmitDevelopmentArtifcts(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/control-plane/kcp-development-artifacts.yaml")
	// THEN
	require.NoError(t, err)

	job := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/control-plane"}), "post-master-kcp-development-artifacts", "master")
	require.NotNil(t, job)
	assert.Empty(t, job.RunIfChanged)
	tester.AssertThatHasExtraRefTestInfra(t, job.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, preset.DockerPushRepoKyma, preset.BuildArtifactsMaster, "preset-kyma-development-artifacts-bucket", preset.GcrPush)
	require.Len(t, job.Spec.Containers, 1)
	cont := job.Spec.Containers[0]
	assert.Equal(t, tester.ImageGolangToolboxLatest, cont.Image)
	require.Len(t, cont.Command, 1)
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-kcp-development-artifacts.sh", cont.Command[0])
	require.Len(t, job.Spec.Volumes, 1)
	assert.Equal(t, "sa-kyma-artifacts", job.Spec.Volumes[0].Name)
	require.NotNil(t, job.Spec.Volumes[0].Secret)
	assert.Equal(t, "sa-kyma-artifacts", job.Spec.Volumes[0].Secret.SecretName)
	require.Len(t, cont.VolumeMounts, 1)
	assert.Equal(t, "sa-kyma-artifacts", cont.VolumeMounts[0].Name)
	assert.Equal(t, "/etc/credentials/sa-kyma-artifacts", cont.VolumeMounts[0].MountPath)
}
