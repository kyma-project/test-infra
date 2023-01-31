package pjconfigtests_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Prowjob,", func() {
	for _, pj := range presubmitFixtures {
		pj := pj
		It("has pubsub config,", func() {
			Expect(pj.Labels).To(MatchFields(IgnoreExtras, Fields{
				"prow.k8s.io/pubsub.project": Equal("sap-kyma-prow"),
				"prow.k8s.io/pubsub.runID":   Not(BeZero()),
				"prow.k8s.io/pubsub.topic":   Equal("prowjobs"),
			}))
		})
	}
})
