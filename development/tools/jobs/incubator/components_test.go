package incubator

import (
	"path"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

const (
	jobBasePath = "./../../../../prow/jobs/incubator"
)

var components = []struct {
	name              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{
		name:  "connector",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "director",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "gateway",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "healthchecker",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "provisioner",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "kyma-environment-broker",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "schema-migrator",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name:  "connectivity-adapter",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release110),
		},
	},
	{
		name:  "pairing-adapter",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release111),
		},
	},
	{
		name:  "external-services-mock",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Optional(),
			jobsuite.Since(releases.Release111),
		},
	},
}

func TestComponentJobs(t *testing.T) {
	testedConfigurations := make(map[string]struct{})
	repos := map[string]struct{}{}
	for _, component := range components {
		t.Run(component.name, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.CompassComponent(component.name, component.image),
				jobsuite.KymaRepo(),
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
	t.Run("All Files covered by test", jobsuite.CheckFilesAreTested(repos, testedConfigurations, jobBasePath, "components"))
}
