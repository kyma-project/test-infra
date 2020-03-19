package incubator

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var tests = []struct {
	name              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{
		name:  "connector-tests",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "director",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic-approach"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release110),
		},
	},
	{
		name:  "provisioner-tests",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "e2e/provisioning",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("test-generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release110),
		},
	},
	{
		name:  "connectivity-adapter",
		image: tester.ImageBootstrap20181204,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.CompassRepo(),
			jobsuite.Since(releases.Release111),
		},
	},
}

func TestTestJobs(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.CompassTest(test.name, test.image),
				jobsuite.KymaRepo(),
				jobsuite.AllReleases(),
			}
			opts = append(opts, test.additionalOptions...)
			cfg := jobsuite.NewConfig(opts...)
			suite := test.suite
			if suite == nil {
				suite = tester.NewComponentSuite
			}
			suite(cfg).Run(t)
		})
	}
}
