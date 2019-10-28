package jobsuite

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
)

type Option func(suite *Config)

func Component(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("components/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

func CompassComponent(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("compass/components/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

func CompassTest(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("compass/tests/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

func Test(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("tests/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

func Tool(name, image string) Option {
	return func(suite *Config) {
		suite.Path = fmt.Sprintf("tools/%s", name)
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

func Project(name, image string) Option {
	return func(suite *Config) {
		suite.Path = name
		suite.Image = image
		suite.FilesTriggeringJob = []string{fmt.Sprintf("%s/fix", suite.Path)}
	}
}

func KymaRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/kyma"
		suite.DockerRepositoryPreset = preset.DockerPushRepoKyma
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

func CompassRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-incubator/compass"
		suite.DockerRepositoryPreset = preset.DockerPushRepoIncubator
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

func TestInfraRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/test-infra"
		suite.DockerRepositoryPreset = preset.DockerPushRepoTestInfra
		suite.BuildPresetMaster = preset.BuildMaster
	}
}

func ConsoleRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/console"
		suite.DockerRepositoryPreset = preset.DockerPushRepoKyma
		suite.BuildPresetMaster = preset.BuildConsoleMaster
	}
}

func DockerRepositoryPreset(preset preset.Preset) Option {
	return func(suite *Config) {
		suite.DockerRepositoryPreset = preset
	}
}

func JobFileSuffix(suffix string) Option {
	return func(suite *Config) {
		suite.JobsFileSuffix = "-" + suffix
	}
}

func Until(rel *releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesUntil(rel)
		suite.Deprecated = true
	}
}

func AllReleases() Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesUntil(releases.Release15)
	}
}

func Since(rel *releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesBetween(rel, releases.Release15)
	}
}

func RunIfChanged(filesTriggeringJob ...string) Option {
	return func(suite *Config) {
		suite.FilesTriggeringJob = filesTriggeringJob
	}
}

func Optional() Option {
	return func(suite *Config) {
		suite.Optional = true
	}
}

func PatchReleases(patchReleases ...*releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.PatchReleases = patchReleases
	}
}
