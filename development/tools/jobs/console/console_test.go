package console

import (
	"path"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

const (
	jobBasePath = "./../../../../prow/jobs/"
)

var loggingYamlName = "console-logging"

var components = []struct {
	path              string
	image             string
	yamlName          *string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{path: "core-ui", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
	{path: "core", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
	{path: "service-catalog-ui", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
	{path: "add-ons", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
	{path: "logging", yamlName: &loggingYamlName, image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
	{path: "tests", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
	{path: "content", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
		},
	},
}

func TestConsoleJobs(t *testing.T) {
	testedConfigurations := make(map[string]struct{})
	repos := map[string]struct{}{}
	for _, component := range components {
		t.Run(component.path, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.Project(component.path, component.yamlName, component.image),
				jobsuite.ConsoleRepo(),
				jobsuite.AllReleases(),
			}

			opts = append(opts, component.additionalOptions...)

			cfg := jobsuite.NewConfig(opts...)
			suite := component.suite
			if suite == nil {
				suite = tester.NewComponentSuite
			}
			ts := suite(cfg)

			if pathProvider, ok := ts.(jobsuite.JobConfigPathProvider); ok {
				testedConfigurations[path.Clean(pathProvider.JobConfigPath())] = struct{}{}
			}
			repos[cfg.Repository] = struct{}{}
			ts.Run(t)
		})
	}
	// t.Run("All Files covered by test", jobsuite.CheckFilesAreTested(repos, testedConfigurations, jobBasePath, "components"))
}
