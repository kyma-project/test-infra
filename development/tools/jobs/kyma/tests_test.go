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
	{path: "service-catalog", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "application-connector-tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "application-gateway-tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "application-operator-tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "application-registry-tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "compass-runtime-agent", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "connection-token-handler-tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "connector-service-tests", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "console-backend-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.AllReleases(),
			jobsuite.RunIfChanged("components/console-backend-service/main.go", "scripts/go-dep.mk"),
		},
	},
	{path: "end-to-end/upgrade", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite, additionalOptions: []jobsuite.Option{
		jobsuite.RunIfChanged("^tests/end-to-end/upgrade/[^chart]", "tests/end-to-end/upgrade/fix"),
		jobsuite.JobFileSuffix("tests-generic"),
		jobsuite.AllReleases(),
	}},
	{path: "end-to-end/external-solution-integration", image: tester.ImageGolangBuildpack1_11,
		additionalOptions: []jobsuite.Option{
			jobsuite.Until(releases.Release19),
		},
	},
	{path: "end-to-end/external-solution-integration", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release110),
		},
	},
	{path: "end-to-end/kubeless-integration", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "event-bus", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "integration/event-service", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "kubeless", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "integration/apiserver-proxy", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "integration/cluster-users", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "integration/dex", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "integration/logging", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "integration/monitoring", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.AllReleases(),
		},
	},
	{path: "knative-serving", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release19),
		},
	},
	{path: "contract/knative-channel-kafka", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("tests-generic"),
			jobsuite.Since(releases.Release110),
		},
	},
	{path: "end-to-end/backup", image: tester.ImageBootstrap20181204, suite: tester.NewGenericComponentSuite,
		additionalOptions: []jobsuite.Option{
			jobsuite.JobFileSuffix("generic"),
			jobsuite.Since(releases.Release110),
		},
	},
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
