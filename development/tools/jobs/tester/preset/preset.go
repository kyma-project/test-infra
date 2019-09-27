package preset

// Preset represents a existing presets
type Preset string

const (
	// WebsiteBotZenHubToken means zenhub token
	WebsiteBotZenHubToken Preset = "preset-website-bot-zenhub-token"
	// KindVolumesMounts means kubernetes-in-docker preset
	KindVolumesMounts Preset = "preset-kind-volume-mounts"
	// GcrPush means GCR push service account
	GcrPush Preset = "preset-sa-gcr-push"
	// DockerPushRepo means Docker repository
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
	// BuildConsoleMaster means console master environment
	BuildConsoleMaster Preset = "preset-build-console-master"
	// BuildConsoleMaster means console PR environment
	BuildConsolePr Preset = "preset-build-console-pr"
	// BuildRelease means release environment
	BuildRelease Preset = "preset-build-release"
	// BotGithubToken means github token
	BotGithubToken Preset = "preset-bot-github-token"
	// BotGithubSSH means github ssh
	BotGithubSSH Preset = "preset-bot-github-ssh"
	// BotGithubIdentity means github identity
	BotGithubIdentity Preset = "preset-bot-github-identity"
	// WebsiteBotGithubToken means github token
	WebsiteBotGithubToken Preset = "preset-website-bot-github-token"
	// KymaGuardBotGithubToken represents the Kyma Guard Bot token for GitHub
	KymaGuardBotGithubToken Preset = "preset-kyma-guard-bot-github-token"
	// WebsiteBotGithubSSH means github ssh
	WebsiteBotGithubSSH Preset = "preset-website-bot-github-ssh"
	// WebsiteBotGithubIdentity means github identity
	WebsiteBotGithubIdentity Preset = "preset-website-bot-github-identity"
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
)
