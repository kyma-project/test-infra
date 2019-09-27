package testinfra

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

func TestWatchPods(t *testing.T) {
	config := jobsuite.NewConfig(
		jobsuite.TestInfraRepo(),
		jobsuite.Project("watch-pods", tester.ImageGolangBuildpack1_11),
		jobsuite.AllReleases(),
		jobsuite.DockerRepositoryPresetSuffix("kyma"),
	)
	tester.NewComponentSuite(config).Run(t)
}
