# TODO:
# These actions variables must exist in github in order to let terraform executor run.
# They are required for getting terraform executor github personal access token from gcp secret manager.
# We must develop a solution for creating a minimal initial setup to let terraform executor apply our whole config.

# TODO(dekiel): Another GitHub variables related to workload identity federation are defined in gcp-workload-identity-federation.tf file.

# GCP project id that contains secret manager with secrets
resource "github_actions_organization_variable" "gcp_kyma_project_project_id" {
  provider      = github.kyma_project
  visibility    = "all"
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


# Name of the secret manager's secret holding kyma bot token for plan prod terraform workflow.
resource "github_actions_variable" "github_terraform_planner_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_PLANNER_SECRET_NAME"
  value         = "kyma-bot-gh-com-terraform-planner-token"
}

data "github_organization" "kyma-project" {
  provider = github.kyma_project
  name     = "kyma-project"
}