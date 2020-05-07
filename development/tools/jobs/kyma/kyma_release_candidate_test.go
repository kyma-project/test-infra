package kyma_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

func TestKymaReleaseCandidateJobsPostsubmit(t *testing.T) {
	for _, currentRelease := range releases.GetAllKymaReleases() {
		t.Run(currentRelease.String(), func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-release-candidate.yaml")
			require.NoError(t, err)

			actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.AllStaticPostsubmits([]string{"kyma-project/kyma"}), tester.GetReleasePostSubmitJobName("kyma-release-candidate", currentRelease), currentRelease.Branch())
			require.NotNil(t, actualJob)

			// then
			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			assert.True(t, actualJob.Decorate)
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, currentRelease.Branch())
			assert.Equal(t, tester.ImageKymaIntegrationK15, actualJob.Spec.Containers[0].Image)
			assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-release-candidate.sh"}, actualJob.Spec.Containers[0].Args)
			tester.AssertThatHasPresets(t, actualJob.JobBase, preset.DindEnabled, "preset-kyma-artifacts-bucket")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_COMPUTE_ZONE", "europe-west4-a")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_COMPUTE_REGION", "europe-west4")
			assert.Equal(t, "true", actualJob.JobBase.Labels["preset-dind-enabled"])
			assert.Equal(t, "true", actualJob.JobBase.Labels["preset-kyma-artifacts-bucket"])
			assert.Equal(t, "true", actualJob.JobBase.Labels["preset-sa-gke-kyma-integration"])
			assert.Equal(t, "true", actualJob.JobBase.Labels["preset-gc-project-env"])
			assert.Equal(t, "true", actualJob.JobBase.Labels["preset-gke-kyma-developers-group"])
		})
	}
}
