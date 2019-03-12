package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndBackupRstoreTestReleases(t *testing.T) {
	// WHEN
	unsupportedReleases := []string{"release-0.6", "release-0.7"}

	for _, currentRelease := range tester.GetSupportedReleases(unsupportedReleases) {
		t.Run(currentRelease, func(t *testing.T) {
			jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/backup-restore-test/backup-restore-test.yaml")
			// THEN
			require.NoError(t, err)
			actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], tester.GetReleaseJobName("kyma-tests-end-to-end-backup-restore-test", currentRelease), currentRelease)
			require.NotNil(t, actualPresubmit)
			assert.False(t, actualPresubmit.SkipReport)
			assert.True(t, actualPresubmit.Decorate)
			assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
			tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease)
			tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildRelease)
			assert.True(t, actualPresubmit.AlwaysRun)
			tester.AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, tester.ImageGolangBuildpack1_11, "/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test")
		})
	}
}

func TestEndToEndBackupRstoreTestJobsPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/backup-restore-test/backup-restore-test.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/kyma"], "pre-master-kyma-tests-end-to-end-backup-restore-test", "master")
	expName := "pre-master-kyma-tests-end-to-end-backup-restore-test"
	assert.Equal(t, expName, actualPresubmit.Name)
	require.NotNil(t, actualPresubmit)
	assert.False(t, actualPresubmit.SkipReport)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPresubmit.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^tests/end-to-end/backup-restore-test/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, actualPresubmit, "tests/end-to-end/backup-restore-test/fix")
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test"}, actualPresubmit.Spec.Containers[0].Args)

}

func TestEndToEndBackupRstoreTestJobsPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/kyma/tests/end-to-end/backup-restore-test/backup-restore-test.yaml")
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/kyma"], "post-master-tests-end-to-end-backup-restore-test", "master")
	expName := "post-master-tests-end-to-end-backup-restore-test"
	assert.Equal(t, expName, actualPost.Name)
	require.NotNil(t, actualPost)

	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, "github.com/kyma-project/kyma", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepo, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^tests/end-to-end/backup-restore-test/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageGolangBuildpack1_11, actualPost.Spec.Containers[0].Image)
	tester.AssertThatHasCommand(t, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test"}, actualPost.Spec.Containers[0].Args)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, actualPost.Spec.Containers[0].Command)
}
