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

resource "github_actions_variable" "gh_tools_kyma_prow_bot_token_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = var.gh_tools_kyma_prow_bot_token_secret_name
  value         = google_secret_manager_secret.kyma_modules_runtime_internal_github_token.secret_id
}

# ==============================================================================
# neighbors-gh-automations — GitHub App credentials + WIF access
# ==============================================================================

variable "neighbors_gh_automations_internal_app_id_secret_id" {
  type        = string
  default     = "neighbors-gh-automations_internal-github-app-id"
  description = "GCP Secret Manager secret ID for the neighbors-gh-automations GitHub App ID on the internal GitHub (github.tools.sap)"
}

variable "neighbors_gh_automations_internal_private_key_secret_id" {
  type        = string
  default     = "neighbors-gh-automations_internal-github-app-private-key"
  description = "GCP Secret Manager secret ID for the neighbors-gh-automations GitHub App private key on the internal GitHub (github.tools.sap)"
}

variable "gh_neighbors_automations_internal_app_id_secret_var_name" {
  type        = string
  default     = "GH_NEIGHBORS_AUTOMATIONS_INTERNAL_APP_ID_SECRET_NAME"
  description = "GitHub Actions repository variable name that holds the GCP secret name for the neighbors-gh-automations App ID"
}

variable "gh_neighbors_automations_internal_private_key_secret_var_name" {
  type        = string
  default     = "GH_NEIGHBORS_AUTOMATIONS_INTERNAL_PRIVATE_KEY_SECRET_NAME"
  description = "GitHub Actions repository variable name that holds the GCP secret name for the neighbors-gh-automations App private key"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager — secret containers
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret" "neighbors_gh_automations_internal_app_id" {
  project   = var.gcp_project_id
  secret_id = var.neighbors_gh_automations_internal_app_id_secret_id

  replication {
    auto {}
  }

  labels = {
    owner           = "neighbors"
    type            = "app-id"
    tool            = "neighbors-gh-automations"
    entity          = "github-app"
    github-instance = "internal"
  }
}

resource "google_secret_manager_secret" "neighbors_gh_automations_internal_private_key" {
  project   = var.gcp_project_id
  secret_id = var.neighbors_gh_automations_internal_private_key_secret_id

  replication {
    auto {}
  }

  labels = {
    owner           = "neighbors"
    type            = "private-key"
    tool            = "neighbors-gh-automations"
    entity          = "github-app"
    github-instance = "internal"
  }
}

# ------------------------------------------------------------------------------
# IAM — grant pull-go-lint reusable workflow secretAccessor via WIF
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "pull_go_lint_neighbors_internal_app_id_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.neighbors_gh_automations_internal_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_go_lint_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "pull_go_lint_neighbors_internal_private_key_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.neighbors_gh_automations_internal_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_go_lint_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# IAM — grant pull-unit-test-go reusable workflow secretAccessor via WIF
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "pull_unit_test_go_neighbors_internal_app_id_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.neighbors_gh_automations_internal_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_unit_test_go_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "pull_unit_test_go_neighbors_internal_private_key_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.neighbors_gh_automations_internal_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_unit_test_go_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# GitHub Actions Repository Variables — expose secret names to reusable workflows
# ------------------------------------------------------------------------------

resource "github_actions_variable" "neighbors_gh_automations_internal_app_id_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = var.gh_neighbors_automations_internal_app_id_secret_var_name
  value         = google_secret_manager_secret.neighbors_gh_automations_internal_app_id.secret_id
}

resource "github_actions_variable" "neighbors_gh_automations_internal_private_key_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = var.gh_neighbors_automations_internal_private_key_secret_var_name
  value         = google_secret_manager_secret.neighbors_gh_automations_internal_private_key.secret_id
}
