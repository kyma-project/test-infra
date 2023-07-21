package pjconfigtests

import (
	"strings"

	"github.com/kyma-project/test-infra/development/opagatekeeper"
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
