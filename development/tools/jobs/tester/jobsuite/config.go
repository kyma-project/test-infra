package jobsuite

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

// Config struct
type Config struct {
	Path                   string
	YamlName               *string
	Repository             string
	Image                  string
	Releases               []*releases.SupportedRelease
	FilesTriggeringJob     []string
	JobsFileSuffix         string
	Deprecated             bool
	DockerRepositoryPreset preset.Preset
	Optional               bool
	BuildPresetMaster      preset.Preset
	PatchReleases          []*releases.SupportedRelease
}

// NewConfig returns new Config
func NewConfig(opts ...Option) *Config {
	suite := &Config{}

	for _, opt := range opts {
		opt(suite)
	}
	return suite
}
