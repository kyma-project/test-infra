package testinfra

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
)

func TestBranchProtection(t *testing.T) {
	actual := readConfig(t)
	t.Run("repository kyma, branch master", func(t *testing.T) {
		masterPolicy, err := actual.GetBranchProtection("kyma-project", "kyma", "master")
		require.NoError(t, err)
		require.NotNil(t, masterPolicy)
		assert.True(t, *masterPolicy.Protect)
		require.NotNil(t, masterPolicy.RequiredStatusChecks)
		assert.Len(t, masterPolicy.RequiredStatusChecks.Contexts, 1)
		assert.Contains(t, masterPolicy.RequiredStatusChecks.Contexts, "license/cla")
	})

	for _, relBranch := range tester.GetAllKymaReleaseBranches() {
		t.Run("repository kyma, branch "+relBranch, func(t *testing.T) {
			p, err := actual.GetBranchProtection("kyma-project", "kyma", relBranch)
			require.NoError(t, err)
			assert.NotNil(t, p)
			assert.True(t, *p.Protect)
			require.NotNil(t, p.RequiredStatusChecks)
			assert.Len(t, p.RequiredStatusChecks.Contexts, 3)
			assert.Contains(t, p.RequiredStatusChecks.Contexts, "license/cla")
			assert.Contains(t, p.RequiredStatusChecks.Contexts, "kyma-integration")
			assert.Contains(t, p.RequiredStatusChecks.Contexts, "kyma-gke-integration")
		})
	}

	t.Run("repository test-infra, branch master", func(t *testing.T) {
		p, err := actual.GetBranchProtection("kyma-project", "test-infra", "master")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.True(t, *p.Protect)
		require.NotNil(t, p.RequiredStatusChecks)
		assert.Len(t, p.RequiredStatusChecks.Contexts, 1)
		assert.Contains(t, p.RequiredStatusChecks.Contexts, "license/cla")
	})

	t.Run("repository website, branch master", func(t *testing.T) {
		masterPolicy, err := actual.GetBranchProtection("kyma-project", "website", "master")
		require.NoError(t, err)
		require.NotNil(t, masterPolicy)
		assert.True(t, *masterPolicy.Protect)
		require.NotNil(t, masterPolicy.RequiredStatusChecks)
		assert.Len(t, masterPolicy.RequiredStatusChecks.Contexts, 1)
		assert.Contains(t, masterPolicy.RequiredStatusChecks.Contexts, "license/cla")
	})

}

func readConfig(t *testing.T) config.Config {
	// WHEN
	f, err := os.Open("../../../../prow/config.yaml")
	// THEN
	require.NoError(t, err)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	cfg := config.Config{}
	require.NoError(t, yaml.Unmarshal(b, &cfg))
	return cfg
}
