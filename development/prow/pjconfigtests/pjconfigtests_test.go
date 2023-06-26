package pjconfigtests_test

import (
	"os"
	"path"

	"github.com/kyma-project/test-infra/development/opagatekeeper"
	kprow "github.com/kyma-project/test-infra/development/prow"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"gopkg.in/yaml.v2"
)

const (
	privilegedContainersConstraintFileOrgPath = "kyma-project/test-infra/prow/cluster/resources/gatekeeper-constraints/workloads/privilegedContainers.yaml"
)

var privilegedContainersConstraint opagatekeeper.K8sPSPPrivilegedContainer

var _ = BeforeSuite(func() {
	// Reade privileged containers constraint from file.
	privilegedContainersConstraintFilePath := path.Join(kprow.OrgDefaultClonePath, privilegedContainersConstraintFileOrgPath)
	privilegedContainersConstraint = opagatekeeper.K8sPSPPrivilegedContainer{}
	privilegedContainersConstraintYaml, err := os.ReadFile(privilegedContainersConstraintFilePath)
	Expect(err).To(BeNil())
	err = yaml.Unmarshal(privilegedContainersConstraintYaml, &privilegedContainersConstraint)
	Expect(err).To(BeNil())
})

var _ = Describe("Prowjob,", func() {
	Context("of presubmit type,", func() {
		for _, pj := range presubmitFixtures {
			pj := pj
			It("has pubsub config,", func() {
				Expect(pj.Labels).To(MatchKeys(IgnoreExtras, Keys{
					"prow.k8s.io/pubsub.project": Equal("sap-kyma-prow"),
					"prow.k8s.io/pubsub.runID":   Not(BeZero()),
					"prow.k8s.io/pubsub.topic":   Equal("prowjobs"),
				}), "Presubmit %s is missing pubsub required labels.", pj.Name)
			})
			It("has ownership annotation", func() {
				missingAnnotations := kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)
				Expect(missingAnnotations).To(BeZero(), "%s\n", missingAnnotations)
			})
			When("using privileged container,", func() {
				It("use allowed image", func() {
					for _, container := range pj.Spec.Containers {
						if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
							Expect(container.Image).To(BeElementOf(privilegedContainersConstraint.Spec.Parameters.ExemptImages), "Presubmit %s is using privileged container with not allowed image %s.", pj.Name, container.Image)
						}
					}
				})
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
				}), "Postsubmit %s is missing pubsub required labels.", pj.Name)
			})
			It("has ownership annotation", func() {
				missingAnnotations := kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)
				Expect(missingAnnotations).To(BeZero(), "%s\n", missingAnnotations)
			})
			When("using privileged container,", func() {
				It("use allowed image", func() {
					for _, container := range pj.Spec.Containers {
						if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
							Expect(container.Image).To(BeElementOf(privilegedContainersConstraint.Spec.Parameters.ExemptImages), "Postsubmit %s is using privileged container with not allowed image %s.", pj.Name, container.Image)
						}
					}
				})
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
				}), "Periodic %s is missing pubsub required labels.", pj.Name)
			})
			It("has ownership annotation", func() {
				missingAnnotations := kprow.CheckRequiredAnnotations(pj.Name, pj.Annotations)
				Expect(missingAnnotations).To(BeZero(), "%s\n", missingAnnotations)
			})
			When("using privileged container,", func() {
				It("use allowed image", func() {
					for _, container := range pj.Spec.Containers {
						if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
							Expect(container.Image).To(BeElementOf(privilegedContainersConstraint.Spec.Parameters.ExemptImages), "Periodic %s is using privileged container with not allowed image %s.", pj.Name, container.Image)
						}
					}
				})
			})
		}
	})
})
