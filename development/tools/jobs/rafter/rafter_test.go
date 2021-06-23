package rafter_test

import (
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testInfraExtraRefSHA = "b973e815bb8124a19a82fe6df722ce174d4a7566"

	rafterJobConfigPath = "./../../../../prow/jobs/rafter/rafter.yaml"
	rafterPathAlias     = "github.com/kyma-project/rafter"

	presetRafterBuildMaster = "preset-rafter-build-main"

	buildScriptCommand = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
	rafterPathArg      = "/home/prow/go/src/github.com/kyma-project/rafter"

	makeCommand        = "make"
	integrationTestArg = "integration-test"
)

var (
	commonPresets     = []preset.Preset{preset.DindEnabled, preset.KindVolumesMounts}
	commonPushPresets = []preset.Preset{preset.GcrPush, preset.DockerPushRepoKyma}

	preBuildPresets           = append(append(commonPresets, commonPushPresets...), preset.BuildPr)
	preIntegrationTestPresets = commonPresets

	postBuildPresets           = append(append(commonPresets, commonPushPresets...), presetRafterBuildMaster)
	postIntegrationTestPresets = commonPresets

	releaseBuildPresets           = append(append(commonPresets, commonPushPresets...), preset.BuildRelease)
	releaseIntegrationTestPresets = commonPresets

	postBranches    = []string{"^master$", "^main$"}
	releaseBranches = []string{"v\\d+\\.\\d+\\.\\d+(?:-.*)?$"}
)

func TestRafterJobsPresubmits(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(rafterJobConfigPath)
	require.NoError(t, err)

	for jobName, actualJob := range map[string]struct {
		presets      []preset.Preset
		containerImg string
		command      string
		args         string
	}{
		"pre-rafter": {
			presets:      preBuildPresets,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      buildScriptCommand,
			args:         rafterPathArg,
		},
		"pre-rafter-integration-test": {
			presets:      preIntegrationTestPresets,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      makeCommand,
			args:         integrationTestArg,
		},
	} {
		t.Run(jobName, func(t *testing.T) {
			// when
			preJob := tester.FindPresubmitJobByName(jobConfig.AllStaticPresubmits([]string{"kyma-project/rafter"}), jobName)
			require.NotNil(t, actualJob)

			assert.False(t, preJob.SkipReport)
			assert.True(t, preJob.AlwaysRun)

			assert.False(t, preJob.Optional)
			assert.Equal(t, rafterPathAlias, preJob.PathAlias)
			assert.Equal(t, 10, preJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfraWithSHA(t, preJob.JobBase.UtilityConfig, "main", testInfraExtraRefSHA)
			tester.AssertThatHasPresets(t, preJob.JobBase, actualJob.presets...)
			assert.Empty(t, preJob.RunIfChanged)

			assert.True(t, *preJob.Spec.Containers[0].SecurityContext.Privileged)

			container := preJob.Spec.Containers[0]
			tester.AssertThatContainerHasEnv(t, container, "GO111MODULE", "on")
			tester.AssertThatContainerHasEnv(t, container, "CLUSTER_VERSION", "1.16")
			assert.Equal(t, "3Gi", container.Resources.Requests.Memory().String())
			assert.Equal(t, "2", container.Resources.Requests.Cpu().String())

			assert.Equal(t, actualJob.containerImg, container.Image)
			assert.Equal(t, []string{actualJob.command}, container.Command)
			assert.Equal(t, []string{actualJob.args}, container.Args)
		})
	}
}

func TestRafterJobsPostsubmits(t *testing.T) {
	jobConfig, err := tester.ReadJobConfig(rafterJobConfigPath)
	require.NoError(t, err)

	for jobName, actualJob := range map[string]struct {
		presets      []preset.Preset
		branches     []string
		containerImg string
		command      string
		args         string
	}{
		"post-rafter": {
			presets:      postBuildPresets,
			branches:     postBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      buildScriptCommand,
			args:         rafterPathArg,
		},
		"release-rafter": {
			presets:      releaseBuildPresets,
			branches:     releaseBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      buildScriptCommand,
			args:         rafterPathArg,
		},
		"post-rafter-integration-test": {
			presets:      postIntegrationTestPresets,
			branches:     postBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      makeCommand,
			args:         integrationTestArg,
		},
		"release-rafter-integration-test": {
			presets:      releaseIntegrationTestPresets,
			branches:     releaseBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      makeCommand,
			args:         integrationTestArg,
		},
	} {
		t.Run(jobName, func(t *testing.T) {
			// when
			postJob := tester.FindPostsubmitJobByName(jobConfig.AllStaticPostsubmits([]string{"kyma-project/rafter"}), jobName)

			// then
			require.NotNil(t, actualJob)
			assert.Equal(t, actualJob.branches, postJob.Branches)

			assert.Equal(t, rafterPathAlias, postJob.PathAlias)
			assert.Equal(t, 10, postJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfraWithSHA(t, postJob.JobBase.UtilityConfig, "main", testInfraExtraRefSHA)
			tester.AssertThatHasPresets(t, postJob.JobBase, actualJob.presets...)
			assert.Empty(t, postJob.RunIfChanged)

			assert.True(t, *postJob.Spec.Containers[0].SecurityContext.Privileged)

			container := postJob.Spec.Containers[0]
			tester.AssertThatContainerHasEnv(t, container, "GO111MODULE", "on")
			tester.AssertThatContainerHasEnv(t, container, "CLUSTER_VERSION", "1.16")
			assert.Equal(t, "3Gi", container.Resources.Requests.Memory().String())
			assert.Equal(t, "2", container.Resources.Requests.Cpu().String())

			assert.Equal(t, actualJob.containerImg, container.Image)
			assert.Equal(t, []string{actualJob.command}, container.Command)
			assert.Equal(t, []string{actualJob.args}, container.Args)
		})
	}
}
