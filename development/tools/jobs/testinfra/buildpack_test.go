package testinfra_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-bootstrap"
	actualPresubmit := tester.FindPresubmitJobByName(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^prow/images/bootstrap/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/bootstrap/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/bootstrap"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBootstrapJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-bootstrap"
	actualPost := tester.FindPostsubmitJobByName(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildRelease)
	assert.Equal(t, "^prow/images/bootstrap/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/bootstrap"}, actualPost.Spec.Containers[0].Args)
}

func TestBootstrapHelmJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-bootstrap-helm"
	actualPresubmit := tester.FindPresubmitJobByName(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^prow/images/bootstrap-helm/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/bootstrap-helm/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/bootstrap-helm"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBootstrapHelmJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-bootstrap-helm"
	actualPost := tester.FindPostsubmitJobByName(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildRelease)
	assert.Equal(t, "^prow/images/bootstrap-helm/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/bootstrap-helm"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackGolangJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-buildpack-golang"
	actualPresubmit := tester.FindPresubmitJobByName(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^prow/images/buildpack-golang/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/buildpack-golang/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackGolangJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-buildpack-golang"
	actualPost := tester.FindPostsubmitJobByName(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildRelease)
	assert.Equal(t, "^prow/images/buildpack-golang/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackGolangKubebuilderJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-buildpack-golang-kubebuilder"
	actualPresubmit := tester.FindPresubmitJobByName(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/buildpack-golang-kubebuilder/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackGolangKubebuilderJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-buildpack-golang-kubebuilder"
	actualPost := tester.FindPostsubmitJobByName(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildRelease)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder"}, actualPost.Spec.Containers[0].Args)
}

func TestKymaClusterInfraPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/test-infra"], "pre-test-infra-kyma-cluster-infra", "master")
	require.NotNil(t, actualPresubmit)

	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatJobRunIfChanged(t, actualPresubmit, "prow/images/kyma-cluster-infra/Dockerfile")
	assert.Len(t, actualPresubmit.Spec.Containers, 1)
	actualContainer := actualPresubmit.Spec.Containers[0]
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181204-a6e79be", actualContainer.Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualContainer.Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/kyma-cluster-infra"}, actualContainer.Args)
	tester.AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
}

func TestKymaClusterInfraPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	actualPostsubmit := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/test-infra"], "post-test-infra-kyma-cluster-infra", "master")
	require.NotNil(t, actualPostsubmit)

	assert.True(t, actualPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPostsubmit.PathAlias)
	tester.AssertThatJobRunIfChanged(t, actualPostsubmit, "prow/images/kyma-cluster-infra/Dockerfile")
	assert.Len(t, actualPostsubmit.Spec.Containers, 1)
	actualContainer := actualPostsubmit.Spec.Containers[0]
	assert.Equal(t, "eu.gcr.io/kyma-project/prow/test-infra/bootstrap:v20181204-a6e79be", actualContainer.Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualContainer.Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/kyma-cluster-infra"}, actualContainer.Args)
	tester.AssertThatSpecifiesResourceRequests(t, actualPostsubmit.JobBase)
}

func TestBuildpackNodeJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-buildpack-node"
	actualPresubmit := tester.FindPresubmitJobByName(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^prow/images/buildpack-node/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/buildpack-node/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackNodeJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-buildpack-node"
	actualPost := tester.FindPostsubmitJobByName(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildRelease)
	assert.Equal(t, "^prow/images/buildpack-node/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node"}, actualPost.Spec.Containers[0].Args)
}

func TestCleanerJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-cleaner"
	actualPresubmit := tester.FindPresubmitJobByName(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"master"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildPr)
	assert.Equal(t, "^prow/images/cleaner/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/cleaner/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/cleaner"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestCleanerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-cleaner"
	actualPost := tester.FindPostsubmitJobByName(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"master"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, tester.PresetDindEnabled, tester.PresetDockerPushRepoTestInfra, tester.PresetGcrPush, tester.PresetBuildRelease)
	assert.Equal(t, "^prow/images/cleaner/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/cleaner"}, actualPost.Spec.Containers[0].Args)
}
