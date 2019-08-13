package kyma_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/buildjob"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

func TestLoadTest(t *testing.T) {
	buildjob.NewSuite(
		buildjob.Tool("load-test", tester.ImageGolangBuildpackLatest),
		buildjob.KymaRepo(),
	).Run(t)
}

