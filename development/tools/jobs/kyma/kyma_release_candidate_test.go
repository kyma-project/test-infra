package kyma_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

func TestKymaReleaseCandidateJobsPostsubmit(t *testing.T) {
	for _, currentRelease := range releases.GetAllKymaReleases() {
		t.Run(currentRelease.String(), func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/kyma-release-candidate.yaml")
			require.NoError(t, err)

			actualJob := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-project/kyma"], tester.GetReleasePostSubmitJobName("kyma-release-candidate", currentRelease), currentRelease.Branch())
			require.NotNil(t, actualJob)

			// then
			assert.Equal(t, "github.com/kyma-project/kyma", actualJob.PathAlias)
			assert.True(t, actualJob.Decorate)
			tester.AssertThatSpecifiesResourceRequests(t, actualJob.JobBase)
			tester.AssertThatHasExtraRefTestInfra(t, actualJob.JobBase.UtilityConfig, currentRelease.Branch())
			assert.Equal(t, tester.ImageBootstrapHelm20181121, actualJob.Spec.Containers[0].Image)
			assert.Equal(t, []string{"-c", "${KYMA_PROJECT_DIR}/test-infra/prow/scripts/cluster-integration/kyma-gke-release-candidate.sh"}, actualJob.Spec.Containers[0].Args)
			tester.AssertThatHasPresets(t, actualJob.JobBase, preset.DindEnabled, "preset-kyma-artifacts-bucket")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "GOOGLE_APPLICATION_CREDENTIALS", "/etc/credentials/sa-kyma-release-candidate/service-account.json")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_DNS_ZONE_NAME", "kymapro-zone")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "KYMA_PROJECT_DIR", tester.KymaProjectDir)
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_COMPUTE_ZONE", "europe-west4-a")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_COMPUTE_REGION", "europe-west4")
			tester.AssertThatContainerHasEnv(t, actualJob.Spec.Containers[0], "CLOUDSDK_CORE_PROJECT", "sap-hybris-sf-playground")
			assert.Equal(t, "sa-kyma-release-candidate", actualJob.Spec.Containers[0].VolumeMounts[0].Name)
			assert.Equal(t, "/etc/credentials/sa-kyma-release-candidate", actualJob.Spec.Containers[0].VolumeMounts[0].MountPath)
			assert.True(t, actualJob.Spec.Containers[0].VolumeMounts[0].ReadOnly)
			assert.Equal(t, "sa-kyma-release-candidate", actualJob.Spec.Volumes[0].Name)
			assert.Equal(t, "sa-kyma-release-candidate", actualJob.Spec.Volumes[0].Secret.SecretName)
		})
	}
}
