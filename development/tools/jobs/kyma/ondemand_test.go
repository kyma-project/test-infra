package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPresubmitOnDemandKymaArtifacts(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-ondemand.yaml")
	// THEN
	require.NoError(t, err)

	job := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-artifacts-ondemand", "master")
	require.NotNil(t, job)

	assert.True(t, job.SkipReport)
	assert.False(t, job.AlwaysRun)
	tester.AssertThatJobRunIfChanged(t, job, "tools/watch-pods")
	tester.AssertThatHasExtraRefTestInfra(t, job.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, job.JobBase, tester.PresetDindEnabled, tester.PresetBuildPr, tester.PresetDockerPushRepo, "preset-kyma-ondemands")
	assert.Len(t, job.Spec.Containers, 1)
	cont := job.Spec.Containers[0]
	assert.Equal(t, tester.ImageBootstrap20181204, cont.Image)
	assert.Len(t, cont.Command, 1)
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-ondemand-kyma-artifacts.sh", cont.Command[0])
	tester.AssertThatContainerHasEnv(t, cont, "KYMA_INSTALLER_PUSH_DIR", "pr/")
}
