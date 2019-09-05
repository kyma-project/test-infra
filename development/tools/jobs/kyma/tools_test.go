package kyma

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/buildjob"
	"testing"
)

var tools = []struct {
	path              string
	image             string
	additionalOptions []buildjob.Option
}{
	{path:"load-test", image: tester.ImageGolangBuildpackLatest},
}

func TestToolsJobs(t *testing.T) {
	for _, test := range tools {
		t.Run(test.path, func(t *testing.T) {
			opts := []buildjob.Option{
				buildjob.Tool(test.path, test.image),
				buildjob.KymaRepo(),
			}
			opts = append(opts, test.additionalOptions...)
			buildjob.NewSuite(opts...).Run(t)
		})
	}
}