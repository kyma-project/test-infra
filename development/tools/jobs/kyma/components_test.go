package kyma

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/buildjob"
	"testing"
)

var components = []struct {
	path              string
	image             string
	additionalOptions []buildjob.Option
}{
	{path: "api-controller", image: tester.ImageGolangBuildpack1_12},
	{path: "apiserver-proxy", image: tester.ImageGolangBuildpack1_12},
	{path: "application-broker", image: tester.ImageGolangBuildpack1_11},
	{path: "application-connectivity-certs-setup-job", image: tester.ImageGolangBuildpackLatest},
	{path: "application-connectivity-validator", image: tester.ImageGolangBuildpackLatest},
	{path: "application-gateway", image: tester.ImageGolangBuildpackLatest},
	{path: "application-operator", image: tester.ImageGolangBuildpackLatest},
	{path: "application-registry", image: tester.ImageGolangBuildpackLatest},
	{path: "asset-metadata-service", image: tester.ImageGolangBuildpack1_11},
	{path: "asset-store-controller-manager", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Since(tester.Release15),
		},
	},
	{path: "asset-store-controller-manager", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Until(tester.Release14),
			buildjob.JobFileSuffix("kubebuilder"),
		},
	},
	{path: "asset-upload-service", image: tester.ImageGolangBuildpack1_11},
	{path: "cms-controller-manager", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Since(tester.Release15),
		},
	},
	{path: "cms-controller-manager", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Until(tester.Release14),
			buildjob.JobFileSuffix("kubebuilder"),
		},
	},
	{path: "cms-services", image: tester.ImageGolangBuildpack1_12},
	{path: "compass-runtime-agent", image: tester.ImageGolangBuildpack1_11},
	{path: "connection-token-handler", image: tester.ImageGolangBuildpackLatest},
	{path: "connectivity-certs-controller", image: tester.ImageGolangBuildpackLatest},
	{path: "connector-service", image: tester.ImageGolangBuildpackLatest},
	{path: "console-backend-service", image: tester.ImageGolangBuildpack1_11},
	{path: "dex-static-user-configurer", image: tester.ImageBootstrapLatest},
	{path: "etcd-tls-setup-job", image: tester.ImageGolangBuildpack1_11},
	{path: "event-bus", image: tester.ImageGolangBuildpack1_11},
	{path: "event-service", image: tester.ImageGolangBuildpack1_11},
	{path: "helm-broker", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Until(tester.Release14),
			buildjob.JobFileSuffix("deprecated"),
		},
	},
	{path: "iam-kubeconfig-service", image: tester.ImageGolangBuildpack1_12},
	{path: "istio-kyma-patch", image: tester.ImageBootstrapLatest},
	{path: "k8s-dashboard-proxy", image: tester.ImageGolangBuildpack1_11},
	{path: "function-controller", image: tester.ImageGolangKubebuilderBuildpackLatest},
	{path: "kubeless-images/nodejs", image: tester.ImageGolangBuildpack1_11},
	{path: "kyma-operator", image: tester.ImageGolangBuildpack1_12},
	{path: "namespace-controller", image: tester.ImageGolangBuildpackLatest},
	{path: "service-binding-usage-controller", image: tester.ImageGolangBuildpack1_11},
	{path: "xip-patch", image: tester.ImageBootstrapLatest},
}

func TestComponentJobs(t *testing.T) {
	for _, component := range components {
		t.Run(component.path, func(t *testing.T) {
			opts := []buildjob.Option{
				buildjob.Component(component.path, component.image),
				buildjob.KymaRepo(),
				buildjob.AllReleases(),

			}
			opts = append(opts, component.additionalOptions...)
			buildjob.NewSuite(opts...).Run(t)
		})
	}
}
