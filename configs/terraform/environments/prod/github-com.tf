# GCP project id that contains secret manager with secrets
resource "github_actions_variable" "gcp_kyma_project_project_id" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GCP_KYMA_PROJECT_PROJECT_ID"
  value         = var.gcp_project_id
}

# Name of the secret manager's secret holding kyma bot token with github variables write permissions
resource "github_actions_variable" "github_terraform_executor_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_EXECUTOR_SECRET_NAME"
  value         = "kyma-bot-gh-com-terraform-executor-token"
}
