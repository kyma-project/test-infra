package jobsuite

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
)

type Config struct {
	Path                         string
	Repository                   string
	Image                        string
	Releases                     []*releases.SupportedRelease
	FilesTriggeringJob           []string
	JobsFileSuffix               string
	Deprecated                   bool
	DockerRepositoryPresetSuffix string
}

func NewConfig(opts ...Option) *Config {
	suite := &Config{}

	for _, opt := range opts {
		opt(suite)
	}
	return suite
}
