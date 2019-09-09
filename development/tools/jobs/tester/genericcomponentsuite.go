package tester

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
	"path"
	"strings"
	"testing"
)

// Designed to check validity of jobs generated from /templates/templates/generic-component.yaml
type GenericComponentSuite struct {
	*jobsuite.Config
}

func NewGenericComponentSuite(config *jobsuite.Config) jobsuite.Suite {
	return &GenericComponentSuite{config}
}

func (s GenericComponentSuite) Run(t *testing.T) {
	s.testRunAgainstEnyBranch(t)

	jobConfig, err := ReadJobConfig(s.jobConfigPath())
	require.NoError(t, err)

	t.Run("presubmit", s.testPresubmitJob(jobConfig))
	t.Run("postsubmit", s.testPostsubmitJob(jobConfig))
}

func (s GenericComponentSuite) testRunAgainstEnyBranch(t *testing.T) {
	require.NotEmpty(t, s.branchesToRunAgainst(), "Jobs are not triggered on any branch. If the component is deprecated remove its job file and this test.")
}

func (s GenericComponentSuite) testPresubmitJob(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		job := FindPresubmitJobByName(jobConfig.Presubmits[s.repositorySectionKey()], s.jobName("pre"))
		require.NotNil(t, job)

		assert.False(t, job.SkipReport, "Must not skip report")
		assert.True(t, job.Decorate, "Must decorate")
		assert.Equal(t, 10, job.MaxConcurrency)
		assert.Equal(t, s.Repository, job.PathAlias)

		for _, branch := range s.branchesToRunAgainst() {
			assert.True(t, job.RunsAgainstBranch(branch), "Must run against branch %s", branch)
		}

		s.assertContainer(t, job.JobBase)
		AssertThatSpecifiesResourceRequests(t, job.JobBase)
		AssertThatHasPresets(t, job.JobBase, PresetDindEnabled, s.presetDockerPushRepository(), PresetGcrPush)
		if !s.isTestInfra() {
			AssertThatHasExtraRefTestInfra(t, job.JobBase.UtilityConfig, "master")
		}

		job.RunsAgainstChanges(s.FilesTriggeringJob)
	}
}

func (s GenericComponentSuite) testPostsubmitJob(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		job := FindPostsubmitJobByName(jobConfig.Postsubmits[s.repositorySectionKey()], s.jobName("post"))
		require.NotNil(t, job, "Job must exists")

		assert.True(t, job.Decorate, "Must decorate")
		assert.Equal(t, 10, job.MaxConcurrency)
		assert.Equal(t, s.Repository, job.PathAlias)

		for _, branch := range s.branchesToRunAgainst() {
			assert.True(t, job.RunsAgainstBranch(branch), "Must run against branch %s", branch)
		}

		s.assertContainer(t, job.JobBase)
		AssertThatSpecifiesResourceRequests(t, job.JobBase)
		AssertThatHasPresets(t, job.JobBase, PresetDindEnabled, s.presetDockerPushRepository(), PresetGcrPush)
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

func (s GenericComponentSuite) jobConfigPath() string {
	return fmt.Sprintf("./../../../../prow/jobs/%s/%s/%s%s.yaml", s.repositoryName(), s.Path, s.componentName(), s.JobsFileSuffix)
}

func (s GenericComponentSuite) repositorySectionKey() string {
	return strings.Replace(s.Repository, "github.com/", "", 1)
}

func (s GenericComponentSuite) jobName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, s.moduleName())
}

func (s GenericComponentSuite) moduleName() string {
	return fmt.Sprintf("%s-%s", s.repositoryName(), strings.Replace(s.Path, "/", "-", -1))
}

func (s GenericComponentSuite) workingDirectory() string {
	return fmt.Sprintf("/home/prow/go/src/%s/%s", s.Repository, s.Path)
}

func (s GenericComponentSuite) isTestInfra() bool {
	return s.Repository == "github.com/kyma-project/test-infra"
}

func (s GenericComponentSuite) presetDockerPushRepository() Preset {
	return Preset(fmt.Sprintf("%s-%s", PresetDockerPushRepo, s.DockerRepositoryPresetSuffix))
}

func (s GenericComponentSuite) branchesToRunAgainst() []string {
	result := make([]string, 0, 1)
	if !s.Deprecated {
		result = append(result, "master")
	}

	for _, rel := range s.Releases {
		result = append(result, rel.Branch())
	}

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
