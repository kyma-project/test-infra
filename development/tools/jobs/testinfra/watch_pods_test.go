package testinfra

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

func TestWatchPods(t *testing.T) {
	config := jobsuite.NewConfig(
		jobsuite.TestInfraRepo(),
		jobsuite.Project("watch-pods", nil, tester.ImageGolangBuildpack1_14),
		jobsuite.AllReleases(),
		jobsuite.DockerRepositoryPreset(preset.DockerPushRepoKyma),
	)
	tester.NewComponentSuite(config).Run(t)
}
