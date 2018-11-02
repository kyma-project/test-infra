package ui_api_layer_test

import (
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"k8s.io/test-infra/prow/config"
	"os"
	"regexp"
	"testing"
)

func TestJobs(t *testing.T) {
	// GIVEN
	f, err := os.Open("jobs.yaml")
	require.NoError(t, err)

	defer f.Close()
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	jobConfig := config.JobConfig{}
	// WHEN
	err = yaml.Unmarshal(b, &jobConfig)
	// THEN
	require.NoError(t, err)


	assert.Len(t, jobConfig.Presubmits, 1)
	presubmits, ex := jobConfig.Presubmits["kyma-project/kyma"]
	assert.True(t, ex)

	master := presubmits[0]
	release := presubmits[1]

	for _, sut := range []config.Presubmit{master,release} {
		assert.Equal(t, sut.Name, sut.Context)
		assert.True(t, sut.Optional)
		assert.True(t, sut.SkipReport)
		assert.True(t, sut.Decorate)

		assert.Len(t, sut.ExtraRefs, 1)
		assert.Equal(t, "test-infra", sut.ExtraRefs[0].Repo)

		assert.Len(t, sut.Spec.Containers,1)

	}

	assert.Equal(t, "prow/kyma/components/ui-api-layer", master.Name)
	assert.Equal(t, master.Name+"/release", release.Name)


	assert.Equal(t, []string{"master"}, master.Branches)
	assert.Equal(t, []string{"^release-\\d+\\.\\d+$"}, release.Branches)

	assert.Equal(t, map[string]string{
		"preset-dind-enabled":           "true",
		"preset-sa-gcr-push":            "true",
		"preset-docker-push-repository": "true",
		"preset-build-pr":               "true",
	}, master.Labels)

	assert.Equal(t, map[string]string{
		"preset-dind-enabled":           "true",
		"preset-sa-gcr-push":            "true",
		"preset-docker-push-repository": "true",
		"preset-build-release":          "true",
	}, release.Labels)


}

func TestRegexp(t *testing.T) {
	runIfChanged := "components/ui-api-layer/"
	fileChanged := "components/ui-api-layer/README.md"

	rx, err := regexp.Compile(runIfChanged)
	require.NoError(t,err)
	assert.True(t,rx.MatchString(fileChanged))
}
