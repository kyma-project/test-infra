package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var tools = []struct {
	path              string
	image             string
	additionalOptions []jobsuite.Option
}{
	{path: "load-test", image: tester.ImageGolangBuildpackLatest},
	{path: "alpine-net", image: tester.ImageGolangBuildpackLatest},
	{path: "backup-plugins", image: tester.ImageGolangBuildpackLatest,
		additionalOptions: []jobsuite.Option{
			jobsuite.Since(releases.Release17),
			jobsuite.Optional(),
		},
	},
}

func TestToolsJobs(t *testing.T) {
	for _, test := range tools {
		t.Run(test.path, func(t *testing.T) {
			opts := []jobsuite.Option{
				jobsuite.Tool(test.path, test.image),
				jobsuite.KymaRepo(),
				jobsuite.AllReleases(),
			}
			opts = append(opts, test.additionalOptions...)
			cfg := jobsuite.NewConfig(opts...)
			tester.ComponentSuite{Config: cfg}.Run(t)
		})
	}
}
