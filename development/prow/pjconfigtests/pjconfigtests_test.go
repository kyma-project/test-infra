package pjconfigtests_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Prowjob,", func() {
	Context("of presubmit type,", func() {
		for _, pj := range presubmitFixtures {
			pj := pj
			It("has pubsub config,", func() {
				Expect(pj.Labels).To(MatchKeys(IgnoreExtras, Keys{
					"prow.k8s.io/pubsub.project": Equal("sap-kyma-prow"),
					"prow.k8s.io/pubsub.runID":   Not(BeZero()),
					"prow.k8s.io/pubsub.topic":   Equal("prowjobs"),
				}), "Presubmit %s is missing pubsub required config.", pj.Name)
			})
		}
	})
	Context("of postsubmit type,", func() {
		for _, pj := range postsubmitFixtures {
			pj := pj
			It("has pubsub config,", func() {
				Expect(pj.Labels).To(MatchKeys(IgnoreExtras, Keys{
					"prow.k8s.io/pubsub.project": Equal("sap-kyma-prow"),
					"prow.k8s.io/pubsub.runID":   Not(BeZero()),
					"prow.k8s.io/pubsub.topic":   Equal("prowjobs"),
				}), "Postsubmit %s is missing pubsub required config.", pj.Name)
			})
		}
	})
	Context("of periodic type,", func() {
		for _, pj := range periodicFixtures {
			pj := pj
			It("has pubsub config,", func() {
				Expect(pj.Labels).To(MatchKeys(IgnoreExtras, Keys{
					"prow.k8s.io/pubsub.project": Equal("sap-kyma-prow"),
					"prow.k8s.io/pubsub.runID":   Not(BeZero()),
					"prow.k8s.io/pubsub.topic":   Equal("prowjobs"),
				}), "Periodic %s is missing pubsub required config.", pj.Name)
			})
		}
	})
})
