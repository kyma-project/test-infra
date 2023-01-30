package pjconfigtests_test

import (
	"testing"

	kprow "github.com/kyma-project/test-infra/development/prow"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/test-infra/prow/config"
)

var (
	presubmitFixtures  []config.Presubmit
	postsubmitFixtures []config.Postsubmit
	peridoicFixtures   []config.Periodic
)

func TestProwjobsConfig(t *testing.T) {
	RegisterFailHandler(Fail)

	var err error
	g := NewGomegaWithT(t)
	presubmitFixtures, postsubmitFixtures, peridoicFixtures, err = kprow.GetRepoProwjobsConfigForProwjob()
	g.Expect(err).To(BeNil())

	RunSpecs(t, "Prowjobs config suite")
}
