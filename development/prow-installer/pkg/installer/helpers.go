package installer

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
)

func AddPrefix(config *config.Config, resourcename string) string {
	if config.Prefix != "" {
		return fmt.Sprintf("%s-%s", config.Prefix, resourcename)
	}
	return resourcename
}
