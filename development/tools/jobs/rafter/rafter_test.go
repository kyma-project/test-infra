package rafter_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testInfraExtraRefSHA = "1eb63a0f829878dda83e67b62416c28c23d71d54"

	rafterJobConfigPath = "./../../../../prow/jobs/rafter/rafter.yaml"
	rafterPathAlias = "github.com/kyma-project/rafter"

	presetRafterBuildMaster = "preset-rafter-build-master"

	buildScriptCommand = "/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"
	rafterPathArg = "/home/prow/go/src/github.com/kyma-project/rafter"

	makeCommand = "make"
	integrationTestArg = "integration-test"
	minIOGatewayTestArg = "minio-gateway-test"
	minIOGatewayMigrationTestArg = "minio-gateway-migration-test"
)
var (
	commonPresets = []preset.Preset{preset.DindEnabled, preset.KindVolumesMounts}
	commonPushPresets = []preset.Preset{preset.GcrPush, preset.DockerPushRepoKyma}
	minIOGCPPresets = []preset.Preset{"preset-rafter-minio-gcs-gateway", "preset-sa-gke-kyma-integration"}
	minIOAzurePresets = []preset.Preset{"preset-rafter-minio-az-gateway", "preset-creds-aks-kyma-integration"}

	preBuildPresets = append(append(commonPresets, commonPushPresets...), preset.BuildPr)
	preIntegrationTestPresets = commonPresets
	preMinIOGCPGatewayPresets = append(commonPresets, minIOGCPPresets...)
	preMinIOAzureGatewayPresets = append(append(commonPresets, minIOAzurePresets...), preset.BuildPr)

	postBuildPresets = append(append(commonPresets, commonPushPresets...), presetRafterBuildMaster)
	postIntegrationTestPresets = commonPresets
	postMinIOGCPGatewayPresets = append(commonPresets, minIOGCPPresets...)
	postMinIOAzureGatewayPresets = append(append(commonPresets, minIOAzurePresets...), presetRafterBuildMaster)

	releaseBuildPresets = append(append(commonPresets, commonPushPresets...), preset.BuildRelease)
	releaseIntegrationTestPresets = commonPresets
	releaseMinIOGCPGatewayPresets = append(commonPresets, minIOGCPPresets...)
	releaseMinIOAzureGatewayPresets = append(append(commonPresets, minIOAzurePresets...), preset.BuildRelease)

	postBranches = []string{"^master$"}
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
			presets: preBuildPresets,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      buildScriptCommand,
			args:         rafterPathArg,
		},
		"pre-rafter-integration-test": {
			presets: preIntegrationTestPresets,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      makeCommand,
			args:         integrationTestArg,
		},
		"pre-rafter-minio-gcs-gateway": {
			presets: preMinIOGCPGatewayPresets,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayTestArg,
		},
		"pre-rafter-minio-az-gateway": {
			presets: preMinIOAzureGatewayPresets,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayTestArg,
		},
		"pre-rafter-minio-gcs-gateway-migration": {
			presets: preMinIOGCPGatewayPresets,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayMigrationTestArg,
		},
		"pre-rafter-minio-az-gateway-migration": {
			presets: preMinIOAzureGatewayPresets,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayMigrationTestArg,
		},
	}{
		t.Run(jobName, func(t *testing.T) {
			// when
			preJob := tester.FindPresubmitJobByName(jobConfig.Presubmits["kyma-project/rafter"], jobName)
			require.NotNil(t, actualJob)

			assert.False(t, preJob.SkipReport)
			assert.True(t, preJob.AlwaysRun)
			assert.True(t, preJob.Decorate)
			assert.False(t, preJob.Optional)
			assert.Equal(t, rafterPathAlias, preJob.PathAlias)
			assert.Equal(t, 10, preJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfraWithSHA(t, preJob.JobBase.UtilityConfig, testInfraExtraRefSHA)
			tester.AssertThatHasPresets(t, preJob.JobBase, actualJob.presets...)
			assert.Empty(t, preJob.RunIfChanged)

			assert.True(t, *preJob.Spec.Containers[0].SecurityContext.Privileged)
			assert.Equal(t, "GO111MODULE", preJob.Spec.Containers[0].Env[0].Name)
			assert.Equal(t, "on", preJob.Spec.Containers[0].Env[0].Value)
			assert.Equal(t, "1536Mi", preJob.Spec.Containers[0].Resources.Requests.Memory().String())
			assert.Equal(t, "800m", preJob.Spec.Containers[0].Resources.Requests.Cpu().String())

			assert.Equal(t, actualJob.containerImg, preJob.Spec.Containers[0].Image)
			assert.Equal(t, []string{actualJob.command}, preJob.Spec.Containers[0].Command)
			assert.Equal(t, []string{actualJob.args}, preJob.Spec.Containers[0].Args)
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
			presets: postBuildPresets,
			branches: postBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      buildScriptCommand,
			args:         rafterPathArg,
		},
		"release-rafter": {
			presets: releaseBuildPresets,
			branches: releaseBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      buildScriptCommand,
			args:         rafterPathArg,
		},
		"post-rafter-integration-test": {
			presets: postIntegrationTestPresets,
			branches: postBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      makeCommand,
			args:         integrationTestArg,
		},
		"release-rafter-integration-test": {
			presets: releaseIntegrationTestPresets,
			branches: releaseBranches,
			containerImg: tester.ImageGolangKubebuilder2BuildpackLatest,
			command:      makeCommand,
			args:         integrationTestArg,
		},
		"post-rafter-minio-gcs-gateway": {
			presets: postMinIOGCPGatewayPresets,
			branches: postBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayTestArg,
		},
		"release-rafter-minio-gcs-gateway": {
			presets: releaseMinIOGCPGatewayPresets,
			branches: releaseBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayTestArg,
		},
		"post-rafter-minio-az-gateway": {
			presets: postMinIOAzureGatewayPresets,
			branches: postBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayTestArg,
		},
		"release-rafter-minio-az-gateway": {
			presets: releaseMinIOAzureGatewayPresets,
			branches: releaseBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayTestArg,
		},
		"post-rafter-minio-gcs-gateway-migration": {
			presets: postMinIOGCPGatewayPresets,
			branches: postBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayMigrationTestArg,
		},
		"release-rafter-minio-gcs-gateway-migration": {
			presets: releaseMinIOGCPGatewayPresets,
			branches: releaseBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayMigrationTestArg,
		},
		"post-rafter-minio-az-gateway-migration": {
			presets: postMinIOAzureGatewayPresets,
			branches: postBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayMigrationTestArg,
		},
		"release-rafter-minio-az-gateway-migration": {
			presets: releaseMinIOAzureGatewayPresets,
			branches: releaseBranches,
			containerImg: tester.ImageKymaClusterInfra20191120,
			command:      makeCommand,
			args:         minIOGatewayMigrationTestArg,
		},
	}{
		t.Run(jobName, func(t *testing.T) {
			// when
			preJob := tester.FindPostsubmitJobByName(jobConfig.Postsubmits["kyma-project/rafter"], jobName)

			// then
			require.NotNil(t, actualJob)
			assert.Equal(t, actualJob.branches, preJob.Branches)

			assert.True(t, preJob.Decorate)
			assert.Equal(t, rafterPathAlias, preJob.PathAlias)
			assert.Equal(t, 10, preJob.MaxConcurrency)
			tester.AssertThatHasExtraRefTestInfraWithSHA(t, preJob.JobBase.UtilityConfig, testInfraExtraRefSHA)
			tester.AssertThatHasPresets(t, preJob.JobBase, actualJob.presets...)
			assert.Empty(t, preJob.RunIfChanged)

			assert.True(t, *preJob.Spec.Containers[0].SecurityContext.Privileged)
			assert.Equal(t, "GO111MODULE", preJob.Spec.Containers[0].Env[0].Name)
			assert.Equal(t, "on", preJob.Spec.Containers[0].Env[0].Value)
			assert.Equal(t, "1536Mi", preJob.Spec.Containers[0].Resources.Requests.Memory().String())
			assert.Equal(t, "800m", preJob.Spec.Containers[0].Resources.Requests.Cpu().String())

			assert.Equal(t, actualJob.containerImg, preJob.Spec.Containers[0].Image)
			assert.Equal(t, []string{actualJob.command}, preJob.Spec.Containers[0].Command)
			assert.Equal(t, []string{actualJob.args}, preJob.Spec.Containers[0].Args)
		})
	}
}
