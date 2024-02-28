# GCP project id that contains secret manager with secrets
resource "github_actions_variable" "gcp_kyma_project_secret_manager_project_id" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GCP_KYMA_PROJECT_SECRET_MANAGER_PROJECT_ID"
  value         = var.gcp_project_id
}

# Name of the secret manager's secret holding kyma bot token with github variables write permissions
resource "github_actions_variable" "kyma_bot_token_secret_manager_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "KYMA_BOT_TOKEN_SECRET_NAME"
  value         = "kyma-bot-secret-gh-com-variables-write-token"
}
