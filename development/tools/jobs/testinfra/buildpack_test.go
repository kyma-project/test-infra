package testinfra_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
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
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
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
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
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
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
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
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackGolangKubebuilder2JobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-buildpack-golang-kubebuilder2"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder2/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/buildpack-golang-kubebuilder2/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder2"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackGolangKubebuilder2JobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-buildpack-golang-kubebuilder2"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder2/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder2"}, actualPost.Spec.Containers[0].Args)
}

func TestKymaClusterInfraPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-project/test-infra"], "pre-test-infra-kyma-cluster-infra", "master")
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

	actualPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-project/test-infra"], "post-test-infra-kyma-cluster-infra", "master")
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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
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

	expName := "post-test-infra-buildpack-node-chromium"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-node-chromium/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node-chromium"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackNodeChromiumPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")

	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-buildpack-node-chromium"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/buildpack-node-chromium/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/buildpack-node-chromium/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node-chromium"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackNodeChromiumPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-buildpack-node"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
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
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
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
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/cleaner/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/cleaner"}, actualPost.Spec.Containers[0].Args)
}

func TestVulnerabilityScannerJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-vulnerability-scanner"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/vulnerability-scanner/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/vulnerability-scanner/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/vulnerability-scanner"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestVulnerabilityScannerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-vulnerability-scanner"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/vulnerability-scanner/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/vulnerability-scanner"}, actualPost.Spec.Containers[0].Args)
}

func TestKubectlJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Presubmits, 1)
	infraPresubmits, ex := jobConfig.Presubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "pre-test-infra-kubectl"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/alpine-kubectl/", actualPresubmit.RunIfChanged)
	tester.AssertThatJobRunIfChanged(t, *actualPresubmit, "prow/images/alpine-kubectl/Dockerfile")
	assert.Equal(t, tester.ImageBootstrapLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/alpine-kubectl"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestKubectlJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.Postsubmits, 1)
	infraPost, ex := jobConfig.Postsubmits["kyma-project/test-infra"]
	assert.True(t, ex)

	expName := "post-test-infra-kubectl"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/alpine-kubectl/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/alpine-kubectl"}, actualPost.Spec.Containers[0].Args)
}
