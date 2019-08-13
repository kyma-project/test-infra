package kyma

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/buildjob"
	"testing"
)

var components = []struct {
	name              string
	image             string
	additionalOptions []buildjob.Option
}{
	{name: "api-controller", image: tester.ImageGolangBuildpack1_12},
	{name: "apiserver-proxy", image: tester.ImageGolangBuildpack1_12},
	{name: "application-broker", image: tester.ImageGolangBuildpack1_11},
	{name: "application-connectivity-certs-setup-job", image: tester.ImageGolangBuildpackLatest},
	{name: "application-connectivity-validator", image: tester.ImageGolangBuildpackLatest},
	{name: "application-gateway", image: tester.ImageGolangBuildpackLatest},
	{name: "application-operator", image: tester.ImageGolangBuildpackLatest},
	{name: "application-registry", image: tester.ImageGolangBuildpackLatest},
	{name: "asset-metadata-service", image: tester.ImageGolangBuildpack1_11},
	{name: "asset-store-controller-manager", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Since(tester.Release15),
		},
	},
	{name: "asset-store-controller-manager", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Until(tester.Release14),
			buildjob.JobFileSuffix("kubebuilder"),
		},
	},
	{name: "asset-upload-service", image: tester.ImageGolangBuildpack1_11},
	{name: "cms-controller-manager", image: tester.ImageGolangKubebuilder2BuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Since(tester.Release15),
		},
	},
	{name: "cms-controller-manager", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Until(tester.Release14),
			buildjob.JobFileSuffix("kubebuilder"),
		},
	},
	{name: "cms-services", image: tester.ImageGolangBuildpack1_12},
	{name: "compass-runtime-agent", image: tester.ImageGolangBuildpack1_11},
	{name: "connection-token-handler", image: tester.ImageGolangBuildpackLatest},
	{name: "connectivity-certs-controller", image: tester.ImageGolangBuildpackLatest},
	{name: "connector-service", image: tester.ImageGolangBuildpackLatest},
	{name: "console-backend-service", image: tester.ImageGolangBuildpack1_11},
	{name: "dex-static-user-configurer", image: tester.ImageBootstrapLatest},
	{name: "etcd-tls-setup-job", image: tester.ImageGolangBuildpack1_11},
	{name: "event-bus", image: tester.ImageGolangBuildpack1_11},
	{name: "event-service", image: tester.ImageGolangBuildpack1_11},
	{name: "helm-broker", image: tester.ImageGolangKubebuilderBuildpackLatest,
		additionalOptions: []buildjob.Option{
			buildjob.Until(tester.Release14),
			buildjob.JobFileSuffix("deprecated"),
		},
	},
	{name: "iam-kubeconfig-service", image: tester.ImageGolangBuildpack1_12},
	{name: "istio-kyma-patch", image: tester.ImageBootstrapLatest},
	{name: "k8s-dashboard-proxy", image: tester.ImageGolangBuildpack1_11},
	{name: "function-controller", image: tester.ImageGolangKubebuilderBuildpackLatest},
	{name: "kubeless-images/nodejs", image: tester.ImageGolangBuildpack1_11},
	{name: "kyma-operator", image: tester.ImageGolangBuildpack1_12},
	{name: "namespace-controller", image: tester.ImageGolangBuildpackLatest},
	{name: "service-binding-usage-controller", image: tester.ImageGolangBuildpack1_11},
	{name: "xip-patch", image: tester.ImageBootstrapLatest},
}

func TestComponentJobs(t *testing.T) {
	for _, component := range components {
		t.Run(component.name, func(t *testing.T) {
			opts := []buildjob.Option{
				buildjob.Component(component.name, component.image),
				buildjob.KymaRepo(),
			}
			opts = append(opts, component.additionalOptions...)
			buildjob.NewSuite(opts...).Run(t)
		})
	}
}
