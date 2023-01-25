package testinfra

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
)

func TestBranchProtection(t *testing.T) {
	actual := readConfig(t)

	var testcases = []struct {
		organization string
		repository   string
		branch       string
		contexts     []string
		approvals    int
	}{
		{"kyma-project", "test-infra", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "kyma", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "website", "main", []string{"license/cla", "netlify/kyma-project/deploy-preview", "tide"}, 1},
		{"kyma-project", "website", "archive-snapshots", []string{"license/cla", "netlify/kyma-project-old/deploy-preview", "tide"}, 1},
		{"kyma-project", "community", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "busola", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "console", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "api-gateway", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "examples", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "addons", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-project", "cli", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "varkes", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "vstudio-extension", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "marketplaces", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "compass", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "documentation-component", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "github-slack-connectors", "main", []string{"license/cla", "tide"}, 1},
		{"kyma-incubator", "kyma-showcase", "main", []string{"license/cla", "tide"}, 1},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("Org: %s, repository: %s, branch: %s", testcase.organization, testcase.repository, testcase.branch), func(t *testing.T) {
			masterPolicy, err := actual.GetBranchProtection(testcase.organization, testcase.repository, testcase.branch, []config.Presubmit{})
			require.NoError(t, err)
			require.NotNil(t, masterPolicy)
			assert.True(t, *masterPolicy.Protect)
			require.NotNil(t, masterPolicy.RequiredPullRequestReviews)
			assert.Equal(t, testcase.approvals, *masterPolicy.RequiredPullRequestReviews.Approvals)
			require.NotNil(t, masterPolicy.RequiredStatusChecks)
			assert.Len(t, masterPolicy.RequiredStatusChecks.Contexts, len(testcase.contexts))
			for _, context := range testcase.contexts {
				assert.Contains(t, masterPolicy.RequiredStatusChecks.Contexts, context)
			}
		})
	}
}

func TestBranchProtectionRelease(t *testing.T) {
	actual := readConfig(t)

	for _, currentRelease := range releases.GetAllKymaReleases() {
		relBranch := currentRelease.Branch()
		t.Run("repository kyma, branch "+relBranch, func(t *testing.T) {
			p, err := actual.GetBranchProtection("kyma-project", "kyma", relBranch, []config.Presubmit{})
			require.NoError(t, err)
			assert.NotNil(t, p)
			assert.True(t, *p.Protect)
			require.NotNil(t, p.RequiredStatusChecks)
			assert.Contains(t, p.RequiredStatusChecks.Contexts, "license/cla")
			assert.Contains(t, p.RequiredStatusChecks.Contexts, "tide")
		})
	}

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
