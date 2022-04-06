package kyma

import (
	"path"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

const (
	jobBasePath = "./../../../../prow/jobs/"
)

var components = []struct {
	path              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{path: "application-connectivity-certs-setup-job", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "compass-runtime-agent", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "console-backend-service", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.RunIfChanged("components/console-backend-service/main.go", "scripts/go-dep.mk"),
		},
	},
	{path: "dex-static-user-configurer", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Until(releases.Release124),
		},
	},
	{path: "iam-kubeconfig-service", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Until(releases.Release124),
		},
	},
	{path: "istio-installer", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "kyma-operator", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "permission-controller", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Until(releases.Release124),
		},
	},
	{path: "service-binding-usage-controller", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "function-controller", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "function-runtimes", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "uaa-activator", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "event-publisher-proxy", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "eventing-controller", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "busola-migrator", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
			jobsuite.Optional(),
		},
	},
	{path: "central-application-gateway", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "central-application-connectivity-validator", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "telemetry-operator", image: tester.ImageGolangBuildpack1_16, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
}

func TestComponentJobs(t *testing.T) {
	testedConfigurations := make(map[string]struct{})
	repos := map[string]struct{}{}
	for _, component := range components {
		t.Run(component.path, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.Component(component.path, component.image),
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
	t.Run("All Files covered by test", jobsuite.CheckFilesAreTested(repos, testedConfigurations, jobBasePath, []string{"components"}))
}
