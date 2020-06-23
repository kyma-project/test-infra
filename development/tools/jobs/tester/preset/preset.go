package preset

// Preset represents a existing presets
type Preset string

const (
	// KindVolumesMounts means kubernetes-in-docker preset
	KindVolumesMounts Preset = "preset-kind-volume-mounts"
	// GcrPush means GCR push service account
	GcrPush Preset = "preset-sa-gcr-push"
	// DockerPushRepoKyma means Docker repository
	DockerPushRepoKyma Preset = "preset-docker-push-repository-kyma"
	// DockerPushRepoTestInfra means Docker repository test-infra images
	DockerPushRepoTestInfra Preset = "preset-docker-push-repository-test-infra"
	// DockerPushRepoIncubator means Docker repository incubator images
	DockerPushRepoIncubator Preset = "preset-docker-push-repository-incubator"
	// DockerPushRepoGlobal means Docker global repository for images
	DockerPushRepoGlobal Preset = "preset-docker-push-repository-global"
	// BuildPr means PR environment
	BuildPr Preset = "preset-build-pr"
	// BuildMaster means master environment
	BuildMaster Preset = "preset-build-master"
	// BuildArtifactsMaster means building artifacts master environment
	BuildArtifactsMaster Preset = "preset-build-artifacts-master"
	// BuildConsoleMaster means console master environment
	BuildConsoleMaster Preset = "preset-build-console-master"
	// BuildConsolePr means console PR environment
	BuildConsolePr Preset = "preset-build-console-pr"
	// BuildRelease means release environment
	BuildRelease Preset = "preset-build-release"
	// BotGithubToken means github token
	BotGithubToken Preset = "preset-bot-github-token"
	// KymaGuardBotGithubToken represents the Kyma Guard Bot token for GitHub
	KymaGuardBotGithubToken Preset = "preset-kyma-guard-bot-github-token"
	// DindEnabled means docker-in-docker preset
	DindEnabled Preset = "preset-dind-enabled"
	// SaGKEKymaIntegration means access to service account capable of creating clusters and related resources
	SaGKEKymaIntegration Preset = "preset-sa-gke-kyma-integration"
	// GCProjectEnv means project name is injected as env variable
	GCProjectEnv Preset = "preset-gc-project-env"
	// KymaBackupRestoreBucket means the bucket used for backups and restore in Kyma
	KymaBackupRestoreBucket Preset = "preset-kyma-backup-restore-bucket"
	// KymaBackupCredentials means the credentials for the service account
	KymaBackupCredentials Preset = "preset-kyma-backup-credentials"
	// GardenerAzureIntegration contains all necessary configuration to deploy on gardener azure from prow
	GardenerAzureIntegration Preset = "preset-gardener-azure-kyma-integration"
	// GardenerGCPIntegration contains all necessary configuration to deploy on gardener GCP from prow
	GardenerGCPIntegration Preset = "preset-gardener-gcp-kyma-integration"
	// GardenerAWSIntegration contains all necessary configuration to deploy on gardener GCP from prow
	GardenerAWSIntegration Preset = "preset-gardener-aws-kyma-integration"
	// KymaCLIStable contains all the configuraion to be able to download the stable master kyma CLI binary
	KymaCLIStable Preset = "preset-kyma-cli-stable"
	// KymaSlackChannel contains the configuration for slack
	KymaSlackChannel Preset = "preset-kyma-slack-channel"
	// SlackBotToken contains the token to use the kyma slack bot
	SlackBotToken Preset = "preset-sap-slack-bot-token"
	// StabilityCheckerSlack contains the information for the stability checker slack account
	StabilityCheckerSlack Preset = "preset-stability-checker-slack-notifications"
	// NightlyGithubIntegration contains the information for nightly clusters
	NightlyGithubIntegration Preset = "preset-nightly-github-integration"
	// KymaKeyring contains the kyma secrets
	KymaKeyring Preset = "preset-kyma-keyring"
	// KymaEncriptionKey contains the kyma cryptographic key
	KymaEncriptionKey Preset = "preset-kyma-encryption-key"
	// SaProwJobResourceCleaner means access to service account capable of cleaning various resources
	SaProwJobResourceCleaner Preset = "preset-sa-prow-job-resource-cleaner"
)
