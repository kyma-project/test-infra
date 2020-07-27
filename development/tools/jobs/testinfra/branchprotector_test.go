package testinfra

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
		{"kyma-project", "kyma", "master", []string{"license/cla"}, 1},
		{"kyma-project", "test-infra", "master", []string{"license/cla", "pre-master-test-infra-vpathguard"}, 1},
		{"kyma-project", "website", "master", []string{"license/cla", "netlify/kyma-project/deploy-preview"}, 1},
		{"kyma-project", "community", "master", []string{"license/cla"}, 1},
		{"kyma-project", "console", "master", []string{"license/cla"}, 1},
		{"kyma-project", "examples", "master", []string{"license/cla"}, 1},
		{"kyma-project", "addons", "master", []string{"license/cla"}, 1},
		{"kyma-project", "cli", "master", []string{"license/cla"}, 1},
		{"kyma-project", "helm-broker", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "varkes", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "vstudio-extension", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "service-catalog-tester", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "gcp-service-broker", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "podpreset-crd", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "marketplaces", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "compass", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "documentation-component", "master", []string{"license/cla"}, 1},
		{"kyma-incubator", "github-slack-connectors", "master", []string{"license/cla"}, 1},
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
			assert.Contains(t, p.RequiredStatusChecks.Contexts, generateStatusCheck("kyma-integration", relBranch))
			assert.Contains(t, p.RequiredStatusChecks.Contexts, generateStatusCheck("kyma-gke-integration", relBranch))
			assert.Contains(t, p.RequiredStatusChecks.Contexts, generateStatusCheck("kyma-gke-upgrade", relBranch))
			assert.Contains(t, p.RequiredStatusChecks.Contexts, generateStatusCheck("kyma-gke-central-connector", relBranch))
			assert.Contains(t, p.RequiredStatusChecks.Contexts, generateStatusCheck("kyma-artifacts", relBranch))
			assert.Contains(t, p.RequiredStatusChecks.Contexts, generateStatusCheck("kyma-installer", relBranch))
		})
	}

}

// status check prefix uses shorten version of release branch, because of that we need to generate the name
func generateStatusCheck(commonJobName, releaseBranch string) string {
	rel := strings.Replace(releaseBranch, "release", "rel", -1)
	rel = strings.Replace(rel, ".", "", -1)
	rel = strings.Replace(rel, "-", "", -1)
	return "pre-" + rel + "-" + commonJobName

}

func readConfig(t *testing.T) config.Config {
	// WHEN
	f, err := os.Open("../../../../prow/branchprotector-config.yaml")
	// THEN
	require.NoError(t, err)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	cfg := config.Config{}
	require.NoError(t, yaml.Unmarshal(b, &cfg))
	return cfg
}
