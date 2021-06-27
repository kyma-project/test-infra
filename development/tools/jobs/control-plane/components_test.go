package controlplane_test

import (
	"path"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

const (
	jobBasePath = "./../../../../prow/jobs"
)

var components = []struct {
	name              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{
		name:  "provisioner",
		image: tester.ImageGolangBuildpack1_16,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.ControlPlaneRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "kyma-environment-broker",
		image: tester.ImageGolangBuildpack1_16,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.ControlPlaneRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "schema-migrator",
		image: tester.ImageGolangBuildpack1_16,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("kcp-generic"),
			jobsuite.ControlPlaneRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "kyma-metrics-collector",
		image: tester.ImageGolangBuildpack1_16,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.ControlPlaneRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "kubeconfig-service",
		image: tester.ImageGolangBuildpack1_16,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.ControlPlaneRepo(),
			jobsuite.AllReleases(),
		},
	},
	{
		name:  "subscription-cleanup-job",
		image: tester.ImageGolangBuildpack1_16,
		suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.ControlPlaneRepo(),
			jobsuite.AllReleases(),
		},
	},
}

func TestComponentJobs(t *testing.T) {
	testedConfigurations := make(map[string]struct{})
	repos := map[string]struct{}{}
	for _, component := range components {
		t.Run(component.name, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.Component(component.name, component.image),
				jobsuite.ControlPlaneRepo(),
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
	t.Run("All Files covered by test", jobsuite.CheckFilesAreTested(repos, testedConfigurations, jobBasePath, []string{"components"}))
}
