package installer

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	"strings"
)

func AddPrefix(config *config.Config, resourcename string) string {
	if config.Prefix != "" {
		return fmt.Sprintf("%s-%s", config.Prefix, resourcename)
	}
	return resourcename
}

func TrimName(resourcename string) string {
	return strings.Trim(resourcename, "-")
}

func CutName(resourcename string) string {
	return fmt.Sprintf("%.30s", resourcename)
}

func FormatName(config *config.Config, resourcename string) string {
	var name string
	name = AddPrefix(config, resourcename)
	name = CutName(name)
	name = TrimName(name)
	return name
}
