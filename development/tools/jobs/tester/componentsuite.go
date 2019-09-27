package tester

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/jobsuite"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
	"path"
	"strings"
	"testing"
)

// Designed to check validity of jobs generated from /templates/templates/component.yaml
type ComponentSuite struct{
	*jobsuite.Config
}

func NewComponentSuite(config *jobsuite.Config) jobsuite.Suite {
	return &ComponentSuite{config}
}

func (s ComponentSuite) Run(t *testing.T) {
	jobConfig, err := ReadJobConfig(s.jobConfigPath())
	require.NoError(t, err)

	expectedNumberOfPresubmits := len(s.Releases)
	if !s.Deprecated {
		expectedNumberOfPresubmits++
	}
	require.Len(t, jobConfig.Presubmits, 1)
	require.Len(t, jobConfig.Presubmits[s.repositorySectionKey()], expectedNumberOfPresubmits)

	if !s.Deprecated {
		require.Len(t, jobConfig.Postsubmits, 1)
		require.Len(t, jobConfig.Postsubmits[s.repositorySectionKey()], 1)
	} else {
		require.Empty(t, jobConfig.Postsubmits)
	}

	require.Empty(t, jobConfig.Periodics)

	if !s.Deprecated {
		t.Run("pre-master", s.preMasterTest(jobConfig))
		t.Run("post-master", s.postMasterTest(jobConfig))
	}
	t.Run("release", s.preReleaseTest(jobConfig))
}

func (s ComponentSuite) preMasterTest(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		job := FindPresubmitJobByNameAndBranch(
			jobConfig.Presubmits[s.repositorySectionKey()],
			s.jobName("pre-master"),
			"master",
		)
		require.NotNil(t, job)

		assert.True(t, job.RunsAgainstBranch("master"))
		assert.False(t, job.SkipReport)
		assert.True(t, job.Decorate)
		assert.Equal(t, s.Optional, job.Optional, "Must be optional: %v", s.Optional)
		assert.Equal(t, 10, job.MaxConcurrency)
		assert.Equal(t, s.Repository, job.PathAlias)
		AssertThatExecGolangBuildpack(t, job.JobBase, s.Image, s.workingDirectory())
		AssertThatSpecifiesResourceRequests(t, job.JobBase)
		if !s.isTestInfra() {
			AssertThatHasExtraRefTestInfra(t, job.JobBase.UtilityConfig, "master")
		}
		AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, s.DockerRepositoryPreset, preset.GcrPush, preset.BuildPr)
		job.RunsAgainstChanges(s.FilesTriggeringJob)
	}
}

func (s ComponentSuite) postMasterTest(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		job := FindPostsubmitJobByNameAndBranch(
			jobConfig.Postsubmits[s.repositorySectionKey()],
			s.jobName("post-master"),
			"master",
		)
		require.NotNil(t, job)

		assert.Equal(t, []string{"^master$"}, job.Branches)
		assert.Equal(t, 10, job.MaxConcurrency)
		assert.True(t, job.Decorate)
		assert.Equal(t, s.Repository, job.PathAlias)
		if !s.isTestInfra() {
			AssertThatHasExtraRefTestInfra(t, job.JobBase.UtilityConfig, "master")
		}
		AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, s.DockerRepositoryPreset, preset.GcrPush, s.BuildPresetMaster)
		job.RunsAgainstChanges(s.FilesTriggeringJob)
		AssertThatExecGolangBuildpack(t, job.JobBase, s.Image, s.workingDirectory())
	}
}

func (s ComponentSuite) preReleaseTest(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		for _, currentRelease := range s.Releases {
			t.Run(currentRelease.String(), func(t *testing.T) {
				job := FindPresubmitJobByNameAndBranch(
					jobConfig.Presubmits[s.repositorySectionKey()],
					GetReleaseJobName(s.moduleName(), currentRelease),
					currentRelease.Branch(),
				)
				require.NotNil(t, job)

				assert.Equal(t, []string{currentRelease.Branch()}, job.Branches)
				assert.False(t, job.SkipReport)
				assert.True(t, job.Decorate)
				assert.Equal(t, 10, job.MaxConcurrency)
				assert.Equal(t, s.Repository, job.PathAlias)
				assert.True(t, job.AlwaysRun)
				AssertThatExecGolangBuildpack(t, job.JobBase, s.Image, s.workingDirectory())
				AssertThatSpecifiesResourceRequests(t, job.JobBase)
				if !s.isTestInfra() {
					AssertThatHasExtraRefTestInfra(t, job.JobBase.UtilityConfig, currentRelease.Branch())
				}
				AssertThatHasPresets(t, job.JobBase, preset.DindEnabled, s.DockerRepositoryPreset, preset.GcrPush, preset.BuildRelease)
				job.RunsAgainstChanges(s.FilesTriggeringJob)
			})
		}
	}
}

func (s ComponentSuite) componentName() string {
	return path.Base(s.Path)
}

func (s ComponentSuite) repositoryName() string {
	return path.Base(s.Repository)
}

func (s ComponentSuite) repositorySectionKey() string {
	return strings.Replace(s.Repository, "github.com/", "", 1)
}

func (s ComponentSuite) moduleName() string {
	return fmt.Sprintf("%s-%s", s.repositoryName(), strings.Replace(s.Path, "/", "-", -1))
}

func (s ComponentSuite) jobConfigPath() string {
	return fmt.Sprintf("./../../../../prow/jobs/%s/%s/%s%s.yaml", s.repositoryName(), s.Path, s.componentName(), s.JobsFileSuffix)
}

func (s ComponentSuite) jobName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, s.moduleName())
}

func (s ComponentSuite) workingDirectory() string {
	return fmt.Sprintf("/home/prow/go/src/%s/%s", s.Repository, s.Path)
}

func (s ComponentSuite) isTestInfra() bool {
	return s.Repository == "github.com/kyma-project/test-infra"
}