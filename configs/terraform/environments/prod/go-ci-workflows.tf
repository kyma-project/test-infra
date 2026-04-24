# ==============================================================================
# Go CI Reusable Workflows - IAM Access
# ==============================================================================
# This configuration grants the generic Go CI reusable workflows from the public
# test-infra repository access to the internal GitHub token stored in GCP Secret Manager.
# These workflows are shared across multiple repositories.
# ==============================================================================

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------

variable "pull_go_lint_reusable_workflow_ref" {
  type        = string
  default     = "kyma-project/test-infra/.github/workflows/pull-go-lint.yaml@refs/heads/main"
  description = "Value of the GitHub OIDC token job_workflow_ref claim for the pull-go-lint reusable workflow in the public test-infra repository"
}

variable "pull_unit_test_go_reusable_workflow_ref" {
  type        = string
  default     = "kyma-project/test-infra/.github/workflows/pull-unit-test-go.yaml@refs/heads/main"
  description = "Value of the GitHub OIDC token job_workflow_ref claim for the pull-unit-test-go reusable workflow in the public test-infra repository"
}

variable "gh_tools_kyma_prow_bot_token_secret_name" {
  type        = string
  default     = "GH_TOOLS_KYMA_PROW_BOT_TOKEN_SECRET_NAME"
  description = "GitHub Actions repository variable name that holds the GCP secret name for the internal GitHub token used by Go CI reusable workflows"
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for Reusable Workflows via WIF
# ------------------------------------------------------------------------------

# Grant the pull-go-lint reusable workflow (public test-infra) access to read the internal GitHub token.
resource "google_secret_manager_secret_iam_member" "pull_go_lint_workflow_internal_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_modules_runtime_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_go_lint_reusable_workflow_ref}"
}

# Grant the pull-unit-test-go reusable workflow (public test-infra) access to read the internal GitHub token.
resource "google_secret_manager_secret_iam_member" "pull_unit_test_go_workflow_internal_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_modules_runtime_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_unit_test_go_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# GitHub Actions Repository Variable (github.com/kyma-project/test-infra)
# ------------------------------------------------------------------------------
# Expose the GCP secret name as a repository-level variable so the reusable
# workflows can resolve the secret at runtime.

import {
  to = github_actions_variable.gh_tools_kyma_prow_bot_token_secret_name
  id = "test-infra:${var.gh_tools_kyma_prow_bot_token_secret_name}"
}

resource "github_actions_variable" "gh_tools_kyma_prow_bot_token_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = var.gh_tools_kyma_prow_bot_token_secret_name
  value         = google_secret_manager_secret.kyma_modules_runtime_internal_github_token.secret_id
}
