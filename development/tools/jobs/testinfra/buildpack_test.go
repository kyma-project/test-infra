package testinfra_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-bootstrap"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/bootstrap/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/bootstrap/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/bootstrap"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBootstrapJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-bootstrap"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/bootstrap/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/bootstrap"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackGolangJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-buildpack-golang"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/buildpack-golang/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/buildpack-golang/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackGolangJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-buildpack-golang"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-golang/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackGolangKubebuilder2JobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-buildpack-golang-kubebuilder2"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder2/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/buildpack-golang-kubebuilder2/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder2"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackGolangKubebuilder2JobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-buildpack-golang-kubebuilder2"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-golang-kubebuilder2/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-golang-kubebuilder2"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackNodeJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-buildpack-node"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/buildpack-node/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/buildpack-node/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackNodeJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-buildpack-node-chromium"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-node-chromium/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node-chromium"}, actualPost.Spec.Containers[0].Args)
}

func TestBuildpackNodeChromiumPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")

	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-buildpack-node-chromium"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/buildpack-node-chromium/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/buildpack-node-chromium/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node-chromium"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestBuildpackNodeChromiumPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-buildpack-node"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/buildpack-node/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/buildpack-node"}, actualPost.Spec.Containers[0].Args)
}

func TestCleanerJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-cleaner"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/cleaner/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/cleaner/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/cleaner"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestCleanerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-cleaner"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/cleaner/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/cleaner"}, actualPost.Spec.Containers[0].Args)
}

func TestVulnerabilityScannerJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-vulnerability-scanner"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/vulnerability-scanner/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/vulnerability-scanner/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/vulnerability-scanner"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestVulnerabilityScannerJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-vulnerability-scanner"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/vulnerability-scanner/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/vulnerability-scanner"}, actualPost.Spec.Containers[0].Args)
}

func TestKubectlJobPresubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PresubmitsStatic, 1)
	infraPresubmits := jobConfig.AllStaticPresubmits([]string{"kyma-project/test-infra"})

	expName := "pre-test-infra-kubectl"
	actualPresubmit := tester.FindPresubmitJobByNameAndBranch(infraPresubmits, expName, "master")
	require.NotNil(t, actualPresubmit)
	assert.Equal(t, expName, actualPresubmit.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPresubmit.Branches)
	assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
	assert.False(t, actualPresubmit.SkipReport)
	assert.True(t, actualPresubmit.Decorate)
	assert.False(t, actualPresubmit.Optional)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, actualPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, "^prow/images/alpine-kubectl/", actualPresubmit.RunIfChanged)
	assert.True(t, tester.IfPresubmitShouldRunAgainstChanges(*actualPresubmit, true, "prow/images/alpine-kubectl/Dockerfile"))
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/alpine-kubectl"}, actualPresubmit.Spec.Containers[0].Args)
}

func TestKubectlJobPostsubmit(t *testing.T) {
	// WHEN
	jobConfig, err := tester.ReadJobConfig("./../../../../prow/jobs/test-infra/buildpack.yaml")
	// THEN
	require.NoError(t, err)

	assert.Len(t, jobConfig.PostsubmitsStatic, 1)
	infraPost := jobConfig.AllStaticPostsubmits([]string{"kyma-project/test-infra"})

	expName := "post-test-infra-kubectl"
	actualPost := tester.FindPostsubmitJobByNameAndBranch(infraPost, expName, "master")
	require.NotNil(t, actualPost)
	assert.Equal(t, expName, actualPost.Name)
	assert.Equal(t, []string{"^master$", "^main$"}, actualPost.Branches)
	assert.Equal(t, 10, actualPost.MaxConcurrency)
	assert.True(t, actualPost.Decorate)
	assert.Equal(t, "github.com/kyma-project/test-infra", actualPost.PathAlias)
	tester.AssertThatHasPresets(t, actualPost.JobBase, preset.DindEnabled, preset.DockerPushRepoTestInfra, preset.GcrPush, preset.BuildRelease)
	assert.Equal(t, "^prow/images/alpine-kubectl/", actualPost.RunIfChanged)
	assert.Equal(t, tester.ImageBootstrapTestInfraLatest, actualPost.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/publish-buildpack.sh"}, actualPost.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/images/alpine-kubectl"}, actualPost.Spec.Containers[0].Args)
}
