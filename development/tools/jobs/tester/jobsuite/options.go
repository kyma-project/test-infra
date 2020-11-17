package jobsuite

import (
	"fmt"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

// Option type
type Option func(suite *Config)

// Component function returns Option type
func Component(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("components/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

// CompassComponent function returns Option type
func CompassComponent(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("compass/components/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

// CompassTest function returns Option type
func CompassTest(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("compass/tests/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

// Test function returns Option type
func Test(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("tests/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

// Tool function returns Option type
func Tool(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("tools/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

// Project function returns Option type
func Project(name string, yamlName *string, image string) Option {
	return func(suite *Config) {
		suite.Path = name
		suite.YamlName = yamlName
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

// KymaRepo function returns Option type
func KymaRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/kyma"
		suite.DockerRepositoryPreset = preset.DockerPushRepoKyma
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

// CompassRepo function returns Option type
func CompassRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-incubator/compass"
		suite.DockerRepositoryPreset = preset.DockerPushRepoIncubator
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

// CompassConsoleRepo function returns Option type
func CompassConsoleRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-incubator/compass-console"
		suite.DockerRepositoryPreset = preset.DockerPushRepoIncubator
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

// ControlPlaneRepo function returns Option type
func ControlPlaneRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/control-plane"
		suite.DockerRepositoryPreset = preset.DockerPushRepoControlPlane
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

// TestInfraRepo function returns Option type
func TestInfraRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/test-infra"
		suite.DockerRepositoryPreset = preset.DockerPushRepoTestInfra
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

// ConsoleRepo function returns Option type
func ConsoleRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/console"
		suite.DockerRepositoryPreset = preset.DockerPushRepoKyma
		suite.BuildPresetMaster = preset.BuildConsoleMaster
	}
}

// DockerRepositoryPreset function returns Option type
func DockerRepositoryPreset(preset preset.Preset) Option {
	return func(suite *Config) {
		suite.DockerRepositoryPreset = preset
	}
}

// JobFileSuffix function returns Option type
func JobFileSuffix(suffix string) Option {
	return func(suite *Config) {
		suite.JobsFileSuffix = "-" + suffix
	}
}

// Until function returns Option type that returns all Kyma releases until provided release.
func Until(rel *releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesUntil(rel)
		suite.Deprecated = !rel.IsNotOlderThan(releases.GetNextKymaRelease())
	}
}

// Between function returns Option type that returns all Kyma releases between provided releases.
func Between(since *releases.SupportedRelease, until *releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesBetween(since, until)
		suite.Deprecated = !until.IsNotOlderThan(releases.GetNextKymaRelease())
	}
}

// AllReleases function returns Option type that returns all Kyma releases.
func AllReleases() Option {
	return func(suite *Config) {
		suite.Releases = releases.GetAllKymaReleases()
	}
}

// Since function returns Option type that returns all Kyma releases since provided releases.
func Since(rel *releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesBetween(rel, releases.GetNextKymaRelease())
	}
}

// RunIfChanged function returns Option type
func RunIfChanged(filesTriggeringJob ...string) Option {
	return func(suite *Config) {
		suite.FilesTriggeringJob = filesTriggeringJob
	}
}

// Optional function returns Option type
func Optional() Option {
	return func(suite *Config) {
		suite.Optional = true
	}
}

// PatchReleases function returns Option type
func PatchReleases(patchReleases ...*releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.PatchReleases = patchReleases
	}
}
