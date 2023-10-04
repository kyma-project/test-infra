package pjconfigtests_test

import (
	kprow "github.com/kyma-project/test-infra/pkg/prow"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/test-infra/prow/config"
)

var (
	presubmitFixtures  []config.Presubmit
	postsubmitFixtures []config.Postsubmit
	periodicFixtures   []config.Periodic
)

func TestProwjobsConfig(t *testing.T) {
	// Skip tests if not running in CI environment. This is to avoid running tests locally as it will fail due to missing environment variables and wrong default paths.
	if os.Getenv("CI") != "true" {
		t.Skip()
	}
	// Skip tests if not running in pull request pjconfigtest prowjob. This is to avoid running tests in golang unit test prowjobs.
	if os.Getenv("JOB_NAME") != "pull-"+os.Getenv("REPO_NAME")+"-pjconfigtest" {
		t.Skip()
	}
	RegisterFailHandler(Fail)

	var err error
	g := NewGomegaWithT(t)

	orgName := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")
	g.Expect(orgName).ToNot(BeZero(), "Variable orgName is zero value.")
	g.Expect(repoName).ToNot(BeZero(), "Variable repoName is zero value.")
	inrepoConfigPath := path.Join(kprow.OrgDefaultClonePath, orgName, repoName)
	// Load prowjobs config to be tested.
	presubmitFixtures, postsubmitFixtures, periodicFixtures, err = kprow.GetProwjobsConfigForProwjob(orgName, repoName, kprow.ProwConfigDefaultClonePath, kprow.JobConfigDefaultClonePath, inrepoConfigPath)
	g.Expect(err).To(BeNil())

	RunSpecs(t, "Prowjobs config suite")
}
