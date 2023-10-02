package pjconfigtests

import (
	"github.com/kyma-project/test-infra/pkg/opagatekeeper"
	"strings"
)

func IsPrivilegedAllowedImage(image string, privilegedContainersConstraint opagatekeeper.K8sPSPPrivilegedContainer) bool {
	for _, exemptImage := range privilegedContainersConstraint.Spec.Parameters.ExemptImages {
		if strings.HasSuffix(exemptImage, "*") {
			prefix := strings.TrimSuffix(exemptImage, "*")
			if strings.HasPrefix(image, prefix) {
				return true
			}
		} else {
			if image == exemptImage {
				return true
			}
		}
	}
	return false
}
