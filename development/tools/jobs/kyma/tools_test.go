package kyma

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
)

var tools = []struct {
	path              string
	image             string
	additionalOptions []jobsuite.Option
}{
	{path: "load-test", image: tester.ImageGolangBuildpackLatest},
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
