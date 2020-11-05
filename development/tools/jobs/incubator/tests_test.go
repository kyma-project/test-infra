package incubator

import (
	"path"
	"testing"

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
		image: tester.ImageBootstrapTestInfraLatest,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.CompassRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "director",
		image: tester.ImageBootstrapTestInfraLatest,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic-approach"),
			jobsuite.CompassRepo(),
		},
	},
	{
		name:  "connectivity-adapter",
		image: tester.ImageBootstrapTestInfraLatest,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.CompassRepo(),
			jobsuite.AllReleases(),
		},
	},
}

func TestTestJobs(t *testing.T) {
	testedConfigurations := make(map[string]struct{})
	repos := map[string]struct{}{}
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
			ts := suite(cfg)
			if pathProvider, ok := ts.(jobsuite.JobConfigPathProvider); ok {
				testedConfigurations[path.Clean(pathProvider.JobConfigPath())] = struct{}{}
			}
			repos[cfg.Repository] = struct{}{}
			ts.Run(t)
		})
	}
	t.Run("All Files covered by test", jobsuite.CheckFilesAreTested(repos, testedConfigurations, jobBasePath, []string{"tests"}))
}
