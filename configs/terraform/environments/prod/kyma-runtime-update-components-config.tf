# ==============================================================================
# Update Components Configuration
# ==============================================================================
# This configuration manages resources required by the kyma-modules "update components"
# reusable workflow.
#
# Resources managed:
# - GCP Secret Manager secret for internal GitHub token
# - GitHub Actions repository variable exposing the GCP secret name
# - IAM permissions (WIF principalSet) to allow the reusable workflow to access the secret
# ==============================================================================

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------

variable "kyma_prow_serviceuser_internal_github_token_gcp_secret_name" {
  type        = string
  default     = "kyma-prow-serviceuser-internal-github-token"
  description = "GCP Secret Manager secret name for internal GitHub kyma-prow-serviceaccount token used by kyma-modules update-components workflow"
}

variable "kyma_prow_serviceuser_internal_github_token_gcp_secret_name_github_repository_variable" {
  type        = string
  default     = "KYMA_PROW_SERVICEUSER_INTERNAL_GITHUB_TOKEN_GCP_SECRET_NAME"
  description = "GitHub Actions repository variable name that holds the GCP secret name"
}

variable "internal_github_kyma_modules_repository_name" {
  type        = string
  default     = "kyma-modules"
  description = "Repository name in internal GitHub Enterprise where the variable should be created"
}

variable "kyma_modules_update_components_reusable_workflow_ref" {
  type        = string
  default     = "kyma/test-infra/.github/workflows/reusable-update-components.yml@refs/heads/main"
  description = "Reference to the test-infra update-components reusable workflow"
}
data "github_repository" "kyma_modules_internal" {
  provider = github.internal_github
  name     = "kyma-modules"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - Internal GitHub Token (kyma-neighbors runtime)
# ------------------------------------------------------------------------------

# kyma-prow-serviceuser-internal-github-token
# This secret stores the Personal Access Token for the kyma-prow service user
# on internal GitHub Enterprise. The token is used by the kyma-modules update-components workflow.
resource "google_secret_manager_secret" "kyma_modules_runtime_internal_github_token" {
  project   = var.gcp_project_id
  secret_id = var.kyma_prow_serviceuser_internal_github_token_gcp_secret_name

  replication {
    auto {}
  }

  labels = {
    type            = "github-token"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "product-kyma-runtime"
    entity          = "kyma-prow-serviceuser"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for Reusable Workflow via WIF
# ------------------------------------------------------------------------------

# Grant the kyma-modules update-components reusable workflow access to read the internal GitHub token.
# The principalSet uses attribute.reusable_workflow_ref to identify the specific reusable workflow.
resource "google_secret_manager_secret_iam_member" "kyma_modules_update_components_workflow_internal_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_modules_runtime_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.repository.id/${data.github_repository.kyma_modules_internal.repo_id}/attribute.reusable_workflow_ref/${var.kyma_modules_update_components_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# GitHub Actions Repository Variable (internal GitHub Enterprise)
# ------------------------------------------------------------------------------
# Expose the GCP secret name as a repository-level variable for kyma/kyma-modules.
resource "github_actions_variable" "kyma_modules_runtime_internal_github_token_gcp_secret_name" {
  provider      = github.internal_github
  repository    = var.internal_github_kyma_modules_repository_name
  variable_name = var.kyma_prow_serviceuser_internal_github_token_gcp_secret_name_github_repository_variable
  value         = google_secret_manager_secret.kyma_modules_runtime_internal_github_token.secret_id
}

