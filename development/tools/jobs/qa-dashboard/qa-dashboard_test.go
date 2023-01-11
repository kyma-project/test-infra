package qadashboard_test

import (
	"path"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var components = []struct {
	path              string
	image             string
	yamlName          *string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{path: "qa-dashboard", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite},
}

func TestQualityDashboardJobs(t *testing.T) {
	testedConfigurations := make(map[string]struct{})
	repos := map[string]struct{}{}
	for _, component := range components {
		t.Run(component.path, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.Project(component.path, component.yamlName, component.image),
				jobsuite.QualityDashboardRepo(),
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
}
