package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var tests = []struct {
	path              string
	image             string
	suite             func(config *jobsuite.Config) jobsuite.Suite
	additionalOptions []jobsuite.Option
}{
	{path: "acceptance", image: tester.ImageGolangBuildpackLatest},
	{path: "application-connector-tests", image: tester.ImageGolangBuildpackLatest},
	{path: "application-gateway-tests", image: tester.ImageGolangBuildpackLatest},
	{path: "application-operator-tests", image: tester.ImageGolangBuildpackLatest},
	{path: "application-registry-tests", image: tester.ImageGolangBuildpackLatest},
	{path: "asset-store", image: tester.ImageGolangBuildpack1_11},
	{path: "compass-runtime-agent", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests"),
		},
	},
	{path: "connection-token-handler-tests", image: tester.ImageGolangBuildpackLatest},
	{path: "connector-service-tests", image: tester.ImageGolangBuildpackLatest},
	{path: "console-backend-service", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests"),
			jobsuite.Until(releases.Release15),
		},
	},
	{path: "console-backend-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release16),
			jobsuite.RunIfChanged("components/console-backend-service/main.go", "scripts/go-dep.mk"),
		},
	},
	{path: "end-to-end/backup-restore-test", image: tester.ImageGolangBuildpack1_11},
	{path: "end-to-end/external-solution-integration", image: tester.ImageGolangBuildpack1_11},
	{path: "end-to-end/kubeless-integration", image: tester.ImageGolangBuildpack1_11},
	{path: "end-to-end/upgrade", image: tester.ImageGolangBuildpack1_11, additionalOptions: []jobsuite.Option{
		jobsuite.RunIfChanged("^tests/end-to-end/upgrade/[^chart]", "tests/end-to-end/upgrade/fix"),
	}},
	{path: "event-bus", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests"),
		},
	},
	{path: "integration/api-controller", image: tester.ImageGolangBuildpack1_12,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests"),
		},
	},
	{path: "integration/apiserver-proxy", image: tester.ImageGolangBuildpack1_12,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests"),
		},
	},
	{path: "integration/cluster-users", image: tester.ImageBootstrapLatest},
	{path: "integration/dex", image: tester.ImageGolangBuildpack1_12},
	{path: "integration/event-service", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests"),
		},
	},
	{path: "integration/logging", image: tester.ImageGolangBuildpackLatest, additionalOptions: []jobsuite.Option{
		jobsuite.Since(releases.Release14),
	}},
	{path: "integration/monitoring", image: tester.ImageGolangBuildpackLatest, additionalOptions: []jobsuite.Option{
		jobsuite.Since(releases.Release14),
	}},
	{path: "knative-build", image: tester.ImageGolangBuildpack1_11},
	{path: "knative-serving", image: tester.ImageGolangBuildpack1_11},
	{path: "kubeless", image: tester.ImageGolangBuildpack1_11},
	{path: "test-namespace-controller", image: tester.ImageGolangBuildpackLatest},
}

func TestTestJobs(t *testing.T) {
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.Test(test.path, test.image),
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
