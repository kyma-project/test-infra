package compassconsole

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const compassConsoleJobPath = "./../../../../../prow/jobs/incubator/compass-console/compass/compass-ui.yaml"

func TestCompassConsoleJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(compassConsoleJobPath)

	// then
	require.NoError(t, err)
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.AllStaticPresubmits([]string{"kyma-incubator/compass-console"}), "pull-compass-console-build", "main")
	require.NotNil(t, actualPresubmit)

	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)

	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.KymaPushImages)
	assert.Equal(t, "eu.gcr.io/sap-kyma-neighbors-dev/image-builder:v20221011-f36fe4783-buildkit", actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/image-builder"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--name=incubator/compass-console", "--config=/config/kaniko-build-config.yaml", "--context=.", "--dockerfile=compass/Dockerfile", "--platform=linux/amd64", "--platform=linux/arm64"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestCompassConsoleJobPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(compassConsoleJobPath)

	// then
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-incubator/compass-console"}), "post-compass-console-build", "main")
	require.NotNil(t, actualPostsubmit)

	assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)

	tester.AssertThatHasPresets(t, actualPostsubmit.JobBase, preset.KymaPushImages)
	assert.Equal(t, "eu.gcr.io/sap-kyma-neighbors-dev/image-builder:v20221011-f36fe4783-buildkit", actualPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/image-builder"}, actualPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"--name=incubator/compass-console", "--config=/config/kaniko-build-config.yaml", "--context=.", "--dockerfile=compass/Dockerfile", "--platform=linux/amd64", "--platform=linux/arm64"}, actualPostsubmit.Spec.Containers[0].Args)
}
