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
				missingAnnotations := kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)
				if missingAnnotations.PjName != "" {
					fmt.Printf("Missing required annotations: %s\n", missingAnnotations)
					// AddReportEntry("Missing required annotations:", missingAnnotations)
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
			It("has ownership annotation", func() {
				missingAnnotations := kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)
				if missingAnnotations.PjName != "" {
					fmt.Printf("Missing required annotations: %s\n", missingAnnotations)
					// AddReportEntry("Missing required annotations:", missingAnnotations)
				}
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
			It("has ownership annotation", func() {
				missingAnnotations := kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)
				if missingAnnotations.PjName != "" {
					fmt.Printf("Missing required annotations: %s\n", missingAnnotations)
					// AddReportEntry("Missing required annotations:", missingAnnotations)
				}
			})
		}
	})
})
