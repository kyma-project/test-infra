package apis

import (
	prowjobsv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects
	// to GroupVersionKinds and back
	// ProwJobs was initialized as external one, see:
	// https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/using_an_external_type.md
	AddToSchemes = append(AddToSchemes, prowjobsv1.AddToScheme)
}
