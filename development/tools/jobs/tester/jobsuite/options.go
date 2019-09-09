package jobsuite

import (
	"fmt"
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

func KymaRepo() Option {
	return func(suite *Config) {
		suite.Repository = "github.com/kyma-project/kyma"
		suite.DockerRepositoryPresetSuffix = "kyma"
	}
}

func DockerRepositoryPresetSuffix(suffix string) Option {
	return func(suite *Config) {
		suite.DockerRepositoryPresetSuffix = suffix
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
		suite.Releases = releases.GetAllKymaReleases()
	}
}

func Since(rel *releases.SupportedRelease) Option {
	return func(suite *Config) {
		suite.Releases = releases.GetKymaReleasesSince(rel)
	}
}

func RunIfChanged(filesTriggeringJob ...string) Option {
	return func(suite *Config) {
		suite.FilesTriggeringJob = filesTriggeringJob
	}
}
