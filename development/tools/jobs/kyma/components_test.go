package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var components = []struct {
	path              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{path: "api-controller", image: tester.ImageGolangBuildpack1_12,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "api-controller", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "apiserver-proxy", image: tester.ImageGolangBuildpack1_12,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "apiserver-proxy", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "application-broker", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "application-connectivity-certs-setup-job", image: tester.ImageGolangBuildpackLatest},
	{path: "application-connectivity-validator", image: tester.ImageGolangBuildpackLatest},
	{path: "application-gateway", image: tester.ImageGolangBuildpackLatest},
	{path: "application-operator", image: tester.ImageGolangBuildpackLatest},
	{path: "application-registry", image: tester.ImageGolangBuildpackLatest},
	{path: "asset-metadata-service", image: tester.ImageGolangBuildpack1_11},
	{path: "asset-store-controller-manager", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release15),
		},
	},
	{path: "asset-store-controller-manager", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release14),
			jobsuite.JobFileSuffix("kubebuilder"),
		},
	},
	{path: "asset-upload-service", image: tester.ImageGolangBuildpack1_11},
	{path: "cms-controller-manager", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release15),
		},
	},
	{path: "cms-controller-manager", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release14),
			jobsuite.JobFileSuffix("kubebuilder"),
		},
	},
	{path: "cms-services", image: tester.ImageGolangBuildpack1_12},
	{path: "compass-runtime-agent", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
			jobsuite.JobFileSuffix("deprecated"),
		},
	},
	{path: "compass-runtime-agent", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release16),
		},
	},
	{path: "connection-token-handler", image: tester.ImageGolangBuildpackLatest},
	{path: "connectivity-certs-controller", image: tester.ImageGolangBuildpackLatest},
	{path: "connector-service", image: tester.ImageGolangBuildpackLatest},
	{path: "console-backend-service", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "console-backend-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release16),
			jobsuite.RunIfChanged("components/console-backend-service/main.go", "scripts/go-dep.mk"),
		},
	},
	{path: "dex-static-user-configurer", image: tester.ImageBootstrapLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "dex-static-user-configurer", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "etcd-tls-setup-job", image: tester.ImageGolangBuildpack1_11},
	{path: "event-bus", image: tester.ImageGolangBuildpack1_11},
	{path: "event-service", image: tester.ImageGolangBuildpack1_11},
	{path: "helm-broker", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release14),
			jobsuite.JobFileSuffix("deprecated"),
		},
	},
	{path: "iam-kubeconfig-service", image: tester.ImageGolangBuildpack1_12,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "iam-kubeconfig-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "istio-kyma-patch", image: tester.ImageBootstrapLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "istio-kyma-patch", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "k8s-dashboard-proxy", image: tester.ImageGolangBuildpack1_11},
	{path: "function-controller", image: tester.ImageGolangKubebuilderBuildpackLatest},
	{path: "kubeless-images/nodejs", image: tester.ImageGolangBuildpack1_11},
	{path: "kyma-operator", image: tester.ImageGolangBuildpack1_12,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "kyma-operator", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "namespace-controller", image: tester.ImageGolangBuildpackLatest},
	{path: "service-binding-usage-controller", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
	{path: "xip-patch", image: tester.ImageBootstrapLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "xip-patch", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release17),
		},
	},
}

func TestComponentJobs(t *testing.T) {
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
			suite(cfg).Run(t)
		})
	}
}
