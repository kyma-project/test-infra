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
	{path: "apiserver-proxy", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "application-broker", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "application-connectivity-certs-setup-job", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-connectivity-validator", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-gateway", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-operator", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "application-registry", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "backup-plugins", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "compass-runtime-agent", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "connection-token-handler", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "connectivity-certs-controller", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "connector-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.AllReleases(),
			jobsuite.JobFileSuffix("generic"),
		},
	},
	{path: "console-backend-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
			jobsuite.RunIfChanged("components/console-backend-service/main.go", "scripts/go-dep.mk"),
		},
	},
	{path: "dex-static-user-configurer", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "etcd-tls-setup-job", image: tester.ImageGolangBuildpack1_11},
	{path: "iam-kubeconfig-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "istio-kyma-patch", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "istio-installer", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release110),
		},
	},
	{path: "k8s-dashboard-proxy", image: tester.ImageGolangBuildpack1_11},
	{path: "kyma-operator", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "permission-controller", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release110),
		},
	},
	{path: "service-binding-usage-controller", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "xip-patch", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "function-controller", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "event-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "event-bus", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "uaa-activator", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release19),
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
	t.Run("All Files covered by test", jobsuite.CheckFilesAreTested(repos, testedConfigurations, jobBasePath, "components"))
}
