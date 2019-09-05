package buildjob

import (
	"fmt"
	. "github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/test-infra/prow/config"
	"path"
	"strings"
	"testing"
)

type Suite struct {
	path                     string
	repository               string
	image                    string
	releases                 []*SupportedRelease
	expectedRunIfChanged     string
	fileExpectedToTriggerJob string
	jobsFileSuffix           string
	expectMasterJobs         bool
}


func NewSuite(opts ...Option) *Suite {
	suite := &Suite{
		releases: GetAllKymaReleases(),
	}
	for _, opt := range opts {
		opt(suite)
	}
	setDefaults(suite)
	return suite
}

func setDefaults(s *Suite) {
	if s.expectedRunIfChanged == "" {
		s.expectedRunIfChanged = fmt.Sprintf("^%s/", s.path)
	}
	if s.fileExpectedToTriggerJob == "" {
		s.fileExpectedToTriggerJob = fmt.Sprintf("%s/fix", s.path)
	}
}

func (s *Suite) componentName() string {
	return path.Base(s.path)
}

func (s *Suite) repositoryName() string {
	return path.Base(s.repository)
}

func (s *Suite) repositorySectionKey() string {
	return strings.Replace(s.repository, "github.com/", "", 1)
}

func (s *Suite) moduleName() string {
	return fmt.Sprintf("%s-%s", s.repositoryName(), strings.Replace(s.path, "/", "-", -1))
}

func (s *Suite) jobConfigPath() string {
	return fmt.Sprintf("./../../../../prow/jobs/%s/%s/%s%s.yaml", s.repositoryName(), s.path, s.componentName(), s.jobsFileSuffix)
}

func (s *Suite) jobName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, s.moduleName())
}

func (s *Suite) workingdirectory() string {
	return fmt.Sprintf("/home/prow/go/src/%s/%s", s.repository, s.path)
}

func (s *Suite) Run(t *testing.T) {
	jobConfig, err := ReadJobConfig(s.jobConfigPath())
	require.NoError(t, err)

	expectedNumberOfPresubmits := len(s.releases)
	if !s.expectMasterJobs {
		expectedNumberOfPresubmits++
	}
	require.Len(t, jobConfig.Presubmits, 1)
	require.Len(t, jobConfig.Presubmits[s.repositorySectionKey()], expectedNumberOfPresubmits)

	if !s.expectMasterJobs {
		require.Len(t, jobConfig.Postsubmits, 1)
		require.Len(t, jobConfig.Postsubmits[s.repositorySectionKey()], 1)
	} else {
		require.Empty(t, jobConfig.Postsubmits)
	}

	require.Empty(t, jobConfig.Periodics)

	if !s.expectMasterJobs {
		t.Run("pre-master", s.preMasterTest(jobConfig))
		t.Run("post-master", s.postMasterTest(jobConfig))
	}
	t.Run("release", s.preReleaseTest(jobConfig))
}

func (s *Suite) preMasterTest(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		actualPresubmit := FindPresubmitJobByName(
			jobConfig.Presubmits[s.repositorySectionKey()],
			s.jobName("pre-master"),
			"master",
		)
		require.NotNil(t, actualPresubmit)

		assert.True(t, actualPresubmit.RunsAgainstBranch("master"))
		assert.False(t, actualPresubmit.SkipReport)
		assert.True(t, actualPresubmit.Decorate)
		assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
		assert.Equal(t, s.repository, actualPresubmit.PathAlias)
		AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, s.image, s.workingdirectory())
		AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
		AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, "master")
		AssertThatHasPresets(t, actualPresubmit.JobBase, PresetDindEnabled, PresetDockerPushRepo, PresetGcrPush, PresetBuildPr)
		AssertThatJobRunIfChanged(t, *actualPresubmit, s.fileExpectedToTriggerJob)
		assert.Equal(t, s.expectedRunIfChanged, actualPresubmit.RunIfChanged)
	}
}

func (s *Suite) postMasterTest(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		actualPostsubmit := FindPostsubmitJobByName(
			jobConfig.Postsubmits[s.repositorySectionKey()],
			s.jobName("post-master"),
			"master",
		)
		require.NotNil(t, actualPostsubmit)

		assert.Equal(t, []string{"^master$"}, actualPostsubmit.Branches)
		assert.Equal(t, 10, actualPostsubmit.MaxConcurrency)
		assert.True(t, actualPostsubmit.Decorate)
		assert.Equal(t, s.repository, actualPostsubmit.PathAlias)
		AssertThatHasExtraRefTestInfra(t, actualPostsubmit.JobBase.UtilityConfig, "master")
		AssertThatHasPresets(t, actualPostsubmit.JobBase, PresetDindEnabled, PresetDockerPushRepo, PresetGcrPush, PresetBuildMaster)
		assert.Equal(t, s.expectedRunIfChanged, actualPostsubmit.RunIfChanged)
		AssertThatExecGolangBuildpack(t, actualPostsubmit.JobBase, s.image, s.workingdirectory())
	}
}

func (s *Suite) preReleaseTest(jobConfig config.JobConfig) func(t *testing.T) {
	return func(t *testing.T) {
		for _, currentRelease := range s.releases {
			t.Run(currentRelease.String(), func(t *testing.T) {
				actualPresubmit := FindPresubmitJobByName(
					jobConfig.Presubmits[s.repositorySectionKey()],
					GetReleaseJobName(s.moduleName(), currentRelease),
					currentRelease.Branch(),
				)
				require.NotNil(t, actualPresubmit)

				assert.Equal(t, []string{currentRelease.Branch()}, actualPresubmit.Branches)
				assert.False(t, actualPresubmit.SkipReport)
				assert.True(t, actualPresubmit.Decorate)
				assert.Equal(t, 10, actualPresubmit.MaxConcurrency)
				assert.Equal(t, s.repository, actualPresubmit.PathAlias)
				assert.True(t, actualPresubmit.AlwaysRun)
				AssertThatExecGolangBuildpack(t, actualPresubmit.JobBase, s.image, s.workingdirectory())
				AssertThatSpecifiesResourceRequests(t, actualPresubmit.JobBase)
				AssertThatHasExtraRefTestInfra(t, actualPresubmit.JobBase.UtilityConfig, currentRelease.Branch())
				AssertThatHasPresets(t, actualPresubmit.JobBase, PresetDindEnabled, PresetDockerPushRepo, PresetGcrPush, PresetBuildRelease)
			})
		}
	}
}
