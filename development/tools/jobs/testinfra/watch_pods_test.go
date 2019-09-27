package testinfra

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
)

func TestWatchPods(t *testing.T) {
	config := jobsuite.NewConfig(
		jobsuite.TestInfraRepo(),
		jobsuite.Project("watch-pods", tester.ImageGolangBuildpack1_11),
		jobsuite.AllReleases(),
		jobsuite.DockerRepositoryPreset(preset.DockerPushRepoKyma),
	)
	tester.NewComponentSuite(config).Run(t)
}
