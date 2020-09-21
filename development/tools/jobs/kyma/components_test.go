package kyma

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

var components = []struct {
	path              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{path: "apiserver-proxy", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-broker", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-connectivity-certs-setup-job", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-connectivity-validator", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-gateway", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-operator", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-registry", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "compass-runtime-agent", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "connection-token-handler", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "connector-service", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "console-backend-service", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.RunIfChanged("components/console-backend-service/main.go", "scripts/go-dep.mk"),
		},
	},
	{path: "dex-static-user-configurer", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "iam-kubeconfig-service", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "istio-kyma-patch", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Optional(),
			jobsuite.Until(releases.Release116),
		},
	},
	{path: "istio-installer", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "k8s-dashboard-proxy", image: tester.ImageGolangBuildpack1_11},
	{path: "kyma-operator", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "permission-controller", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "service-binding-usage-controller", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "xip-patch", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "function-controller", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "function-runtimes", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "event-service", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "nats-init", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
			jobsuite.Optional(),
		},
	},
	{path: "uaa-activator", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "event-sources", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "event-publisher-proxy", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release117),
		},
	},
	{path: "eventing-controller", image: tester.ImageBootstrapTestInfraLatest, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release117),
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
