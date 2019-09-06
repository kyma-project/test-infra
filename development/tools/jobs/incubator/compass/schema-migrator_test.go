package compass_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const schemaMigratorJobPath = "./../../../../../prow/jobs/incubator/compass/components/schema-migrator/schema-migrator.yaml"

func TestSchemaMigratorJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(schemaMigratorJobPath)
	// THEN
	require.NoError(t, err)

	actualPre := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-incubator/compass"], "pre-master-compass-components-schema-migrator", "master")
	require.NotNil(t, actualPre)

	assert.Equal(t, 10, actualPre.MaxConcurrency)
	assert.False(t, actualPre.SkipReport)
	assert.True(t, actualPre.Decorate)
	assert.False(t, actualPre.Optional)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualPre.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPre.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPre.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^components/schema-migrator/", actualPre.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPre, "components/schema-migrator/some_random_file.sh")
	assert.Len(t, actualPre.Spec.Containers, 1)
	testContainer := actualPre.Spec.Containers[0]
	assert.Equal(t, tester.ImageBootstrap20181204, testContainer.Image)
}

func TestSchemaMigratorJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig(schemaMigratorJobPath)
	// THEN
	require.NoError(t, err)

	actualPost := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-incubator/compass"], "post-master-compass-components-schema-migrator", "master")
	require.NotNil(t, actualPost)

	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/compass", actualPost.PathAlias)
	tester.AssertThatHasExtraRefTestInfra(t, actualPost.JobBase.UtilityConfig, "master")
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoIncubator, tester.PresetGcrPush, tester.PresetBuildMaster)
	assert.Equal(t, "^components/schema-migrator/", actualPost.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPost, "components/schema-migrator/some_random_file.sh")
	assert.Len(t, actualPost.Spec.Containers, 1)
	testContainer := actualPost.Spec.Containers[0]
	assert.Equal(t, tester.ImageBootstrap20181204, testContainer.Image)
}
