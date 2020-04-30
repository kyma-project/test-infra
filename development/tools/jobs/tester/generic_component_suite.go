package tester

import (
	"fmt"
	"log"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"

	"github.com/kyma-project/test-infra/development/tools/jobs/releases"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
)

// Designed to check validity of jobs generated from /templates/templates/generic-component.yaml
type GenericComponentSuite struct {
	*jobsuite.Config
}

func NewGenericComponentSuite(config *jobsuite.Config) jobsuite.Suite {
	return &GenericComponentSuite{config}
}

func (s GenericComponentSuite) Run(t *testing.T) {
	s.testRunAgainstAnyBranch(t)

	jobConfig, err := ReadJobConfig(s.JobConfigPath())
	require.NoError(t, err)

	t.Run("presubmit", s.testPresubmitJob(jobConfig))
	t.Run("postsubmit", s.testPostsubmitJob(jobConfig))
}

func (s GenericComponentSuite) testRunAgainstAnyBranch(t *testing.T) {
	require.NotEmpty(t, s.branchesToRunAgainst(), "Jobs are not triggered on any branch. If the component is deprecated remove its job file and this test.")
}

func (s GenericComponentSuite) testPresubmitJob(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		job := FindPresubmitJobByName(jobConfig.PresubmitsStatic[s.repositorySectionKey()], s.jobName("pre"))
		require.NotNil(t, job)

		assert.False(t, job.SkipReport, "Must not skip report")
		assert.True(t, job.Decorate, "Must decorate")
		assert.Equal(t, s.Optional, job.Optional, "Must be optional: %v", s.Optional)
		assert.Equal(t, 10, job.MaxConcurrency)
		assert.Equal(t, s.Repository, job.PathAlias)

		for _, branch := range s.branchesToRunAgainst() {
			assert.True(t, job.CouldRun(branch), "Must run against branch %s", branch)
		}
		for _, branch := range s.branchesNotToRunAgainst() {
			assert.False(t, job.CouldRun(branch), "Must NOT run against branch %s", branch)
		}

		s.assertContainer(t, job.JobBase)
		AssertThatSpecifiesResourceRequests(t, job.JobBase)
		AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, s.DockerRepositoryPreset, preset.GcrPush)
		if !s.isTestInfra() {
			AssertThatHasExtraRefTestInfra(t, job.JobBase.UtilityConfig, "master")
		}

		job.RunsAgainstChanges(s.FilesTriggeringJob)
	}
}

func (s GenericComponentSuite) testPostsubmitJob(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		job := FindPostsubmitJobByName(jobConfig.PostsubmitsStatic[s.repositorySectionKey()], s.jobName("post"))
		require.NotNil(t, job, "Job must exists")

		assert.True(t, job.Decorate, "Must decorate")
		assert.Equal(t, 10, job.MaxConcurrency)
		assert.Equal(t, s.Repository, job.PathAlias)

		for _, branch := range s.branchesToRunAgainst() {
			assert.True(t, job.CouldRun(branch), "Must run against branch %s", branch)
		}

		s.assertContainer(t, job.JobBase)
		AssertThatSpecifiesResourceRequests(t, job.JobBase)
		AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, s.DockerRepositoryPreset, preset.GcrPush)
		if !s.isTestInfra() {
			AssertThatHasExtraRefTestInfra(t, job.JobBase.UtilityConfig, "master")
		}

		job.RunsAgainstChanges(s.FilesTriggeringJob)
	}
}

func (s GenericComponentSuite) componentName() string {
	return path.Base(s.Path)
}

func (s GenericComponentSuite) repositoryName() string {
	return path.Base(s.Repository)
}

func (s GenericComponentSuite) JobConfigPath() string {
	// Components outside kyma-project need this switch, because generic job will create for example:
	// Repository = github.com/kyma-incubator/compass,
	// will generate path: `kyma-incubator` which is not valid in current state
	// Current valid path is `incubator`
	jobConfigPath := ""
	switch {
	case strings.Contains(s.Repository, "kyma-project"):
		jobConfigPath = fmt.Sprintf("./../../../../prow/jobs/%s/%s/%s%s.yaml", s.repositoryName(), s.Path, s.componentName(), s.JobsFileSuffix)

	case strings.Contains(s.Repository, "kyma-incubator"):
		repos := path.Dir(s.Repository)
		org := path.Base(repos)
		orgPath := strings.Replace(org, "kyma-", "", 1)
		jobConfigPath = fmt.Sprintf("./../../../../prow/jobs/%s/%s/%s%s.yaml", orgPath, s.Path, s.componentName(), s.JobsFileSuffix)

	default:
		log.Fatalf("organization not supported: %s", s.Repository)
	}

	return jobConfigPath
}

func (s GenericComponentSuite) repositorySectionKey() string {
	return strings.Replace(s.Repository, "github.com/", "", 1)
}

func (s GenericComponentSuite) jobName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, s.moduleName())
}

func (s GenericComponentSuite) moduleName() string {
	return fmt.Sprintf("%s-%s", s.repositoryName(), strings.Replace(s.getPath(), "/", "-", -1))
}

func (s GenericComponentSuite) workingDirectory() string {
	return fmt.Sprintf("/home/prow/go/src/%s/%s", s.Repository, s.getPath())
}

func (s GenericComponentSuite) getPath() string {
	if strings.Contains(s.Repository, "kyma-project") {
		return s.Path
	}
	// Components outside kyma-project need this workaround
	// Remove first part of Path
	paths := strings.Split(s.Path, "/")
	return strings.Join(paths[1:], "/")
}

func (s GenericComponentSuite) isTestInfra() bool {
	return s.Repository == "github.com/kyma-project/test-infra"
}

func (s GenericComponentSuite) notSupportedComponentReleaseBranches() []string {
	unsupportedBranches := []string{}

	allReleases := releases.GetAllKymaReleases()
FIND:
	for _, rel := range allReleases {
		for _, supportedRelease := range s.Releases {
			if rel.Compare(supportedRelease) == 0 {
				continue FIND
			}
		}
		unsupportedBranches = append(unsupportedBranches, fmt.Sprintf("%v-%v-%v", "release", rel.String(), s.componentName()))
	}
	return unsupportedBranches
}

func (s GenericComponentSuite) componentReleaseBranches() []string {
	releaseBranches := []string{}

	for _, rel := range s.Releases {
		releaseBranches = append(releaseBranches, fmt.Sprintf("%v-%v-%v", "release", rel.String(), s.componentName()))
	}
	return releaseBranches
}

func (s GenericComponentSuite) branchesNotToRunAgainst() []string {
	result := make([]string, 0, 1)
	if s.Deprecated {
		result = append(result, "master")
	}

	unsupportedReleaseBranches := s.notSupportedComponentReleaseBranches()
	result = append(result, unsupportedReleaseBranches...)
	return result
}

func (s GenericComponentSuite) branchesToRunAgainst() []string {
	result := make([]string, 0, 1)
	if !s.Deprecated {
		result = append(result, "master")
	}

	releaseBranches := s.componentReleaseBranches()
	result = append(result, releaseBranches...)
	return result
}

// AssertThatExecGolangBuildpack checks if job executes golang buildpack
func (s GenericComponentSuite) assertContainer(t *testing.T, job config.JobBase) {
	assert.Len(t, job.Spec.Containers, 1, "Must have one container")
	assert.Equal(t, s.Image, job.Spec.Containers[0].Image)
	assert.Len(t, job.Spec.Containers[0].Command, 1, "Must have one command")
	assert.Equal(t, "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build-generic.sh", job.Spec.Containers[0].Command[0])
	assert.Equal(t, []string{s.workingDirectory()}, job.Spec.Containers[0].Args)
	assert.True(t, *job.Spec.Containers[0].SecurityContext.Privileged, "Must run in privileged mode")
}
