package pjconfigtests_test

import (
	"os"
	"path"
	"testing"

	kprow "github.com/kyma-project/test-infra/development/prow"
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
	RegisterFailHandler(Fail)

	var err error
	g := NewGomegaWithT(t)

	orgName := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")
	g.Expect(os.Getenv(orgName)).ToNot(BeZero())
	g.Expect(os.Getenv(repoName)).ToNot(BeZero())
	inrepoConfigPath := path.Join(kprow.OrgDefaultClonePath, orgName, repoName)
	presubmitFixtures, postsubmitFixtures, periodicFixtures, err = kprow.GetProwjobsConfigForProwjob(orgName, repoName, kprow.ProwConfigDefaultClonePath, kprow.JobConfigDefaultClonePath, inrepoConfigPath)
	g.Expect(err).To(BeNil())

	RunSpecs(t, "Prowjobs config suite")
}
