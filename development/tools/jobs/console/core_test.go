package console_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

func TestCore(t *testing.T) {
	config := jobsuite.NewConfig(
		jobsuite.ConsoleRepo(),
		jobsuite.Project("core", tester.ImageNodeChromiumBuildpackLatest),
	)
	tester.NewComponentSuite(config).Run(t)
}
