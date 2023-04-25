package pjconfigtests_test

import (
	"fmt"

	kprow "github.com/kyma-project/test-infra/development/prow"
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
			It("has ownership annotation", func() {
				var errs []error
				errs = append(errs, kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)...)
				if len(errs) > 0 {
					// for now not fail the job, only return information
					// t.Fail()
					for _, e := range errs {
						AddReportEntry(fmt.Sprintf("Prowjob %s is missing required annotations: %s", pj.Name, e.Error()))
					}
				}
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
