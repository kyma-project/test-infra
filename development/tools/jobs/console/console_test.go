package console

import (
	"path"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
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
	{path: "core-ui", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release111),
		},
	},
	{path: "core", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release112),
			jobsuite.JobFileSuffix("tmp"),
			jobsuite.Optional(),
		},
	},
	{path: "service-catalog-ui", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release111),
		},
	},
	{path: "add-ons", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release111),
		},
	},
	{path: "logging", yamlName: &loggingYamlName, image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release111),
		},
	},
	{path: "tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release111),
		},
	},
	{path: "content", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release112),
			jobsuite.JobFileSuffix("tmp"),
			jobsuite.Optional(),
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
