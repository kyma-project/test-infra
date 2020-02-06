package incubator

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var components = []struct {
	name              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{
		name: "connector",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
	{
		name: "director",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
	{
		name: "gateway",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
	{
		name: "healthchecker",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
	{
		name: "provisioner",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
	{
		name: "kyma-environment-broker",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release19),
		},
	},
	{
		name: "schema-migrator",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
	{
		name: "connectivity-adapter",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},

	{
		name: "pairing-adapter",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release18),
		},
	},
}

func TestComponentJobs(t *testing.T) {
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
			suite(cfg).Run(t)
		})
	}
}
