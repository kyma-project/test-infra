# ==============================================================================
# DORA Integration Configuration
# ==============================================================================
# This configuration manages resources for DORA (DevOps Research and Assessment)
# metrics collection and reporting in the kyma project.
#
# Resources managed:
# - GCP Secret Manager secrets for GitHub tokens (public and internal GitHub)
# - GitHub Actions organization variables for secret references
# - IAM permissions for accessing secrets from GitHub Actions workflows
#
# The configuration supports both public GitHub (github.com) and internal
# GitHub Enterprise instances.
# ==============================================================================

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------

variable "dora_integration_gcp_secret_name_internal_github_token" {
  type        = string
  default     = "dora-integration-gh-tools-serviceuser-token"
  description = "GCP Secret Manager secret name for internal GitHub token used by DORA integration"
}

variable "dora_integration_public_github_token_gcp_secret_name_github_organization_variable" {
  type        = string
  default     = "DORA_INTEGRATION_PUBLIC_GITHUB_TOKEN_GCP_SECRET_NAME"
  description = "GitHub Actions variable name for the GCP secret name with public GitHub token"
}

variable "dora_integration_internal_github_token_gcp_secret_name_github_organization_variable" {
  type        = string
  default     = "DORA_INTEGRATION_INTERNAL_GITHUB_TOKEN_GCP_SECRET_NAME"
  description = "GitHub Actions variable name for the GCP secret name with internal GitHub token"
}

variable "dora_integration_reusable_workflow_ref" {
  type        = string
  default     = "kyma-project/test-infra/.github/workflows/dora-integration.yml@refs/heads/main"
  description = "Reference to the DORA integration reusable workflow"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - Internal GitHub Token
# ------------------------------------------------------------------------------

# dora-integration-internal-github-token This secret stores the Personal Access Token for the DORA integration service
# user on internal GitHub. The token is used to collect DORA metrics from internal repositories.
# ------------------------------------------------------------------------------
import {
  id = "projects/${var.gcp_project_id}/secrets/${var.dora_integration_gcp_secret_name_internal_github_token}"
  to = google_secret_manager_secret.dora_integration_internal_github_token
}

resource "google_secret_manager_secret" "dora_integration_internal_github_token" {
  project   = var.gcp_project_id
  secret_id = var.dora_integration_gcp_secret_name_internal_github_token

  replication {
    auto {}
  }

  labels = {
    type            = "github-token"
    tool            = "dora-integration"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "dora-integration-serviceuser"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for GitHub Actions Workflows
# ------------------------------------------------------------------------------

# dora_integration_workflow_public_token_reader Grant the DORA integration reusable workflow access to read the public GitHub token.
# This token (kyma-bot) is used to collect DORA metrics from public repositories in the kyma-project organization on github.com.
# The principalSet uses attribute.reusable_workflow_ref to identify the specific reusable workflow that should have access to these secrets.
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_public_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.dora_integration_reusable_workflow_ref}"
}

# dora_integration_workflow_internal_token_reader Grant the DORA integration reusable workflow access to read the internal GitHub token.
# This token is used to collect DORA metrics from internal repositories in the kyma organization on internal GitHub.
# The principalSet uses attribute.reusable_workflow_ref to identify the specific reusable workflow that should have access to these secrets.
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_internal_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora_integration_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.dora_integration_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# GitHub Actions Organization Variables
# ------------------------------------------------------------------------------
# This section creates organization-level GitHub Actions variables that GitHub action workflows can use.
# ------------------------------------------------------------------------------

# dora-integration-public-github-token-gcp-secret-name Expose the GCP secret name with public GitHub token as a GitHub Actions organization variable.
# This variable will be available to all repositories in the internal GitHub kyma organization.
# The variable is used in internal github workflows to access the public GitHub token from GCP Secret Manager.
# Note: Even though this is for public GitHub tokens, the variable is created
# in internal GitHub organization because that's where the DORA integration workflows run from.
resource "github_actions_organization_variable" "dora_integration_public_github_token_gcp_secret_name" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = var.dora_integration_public_github_token_gcp_secret_name_github_organization_variable
  value         = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
}

# dora-integration-internal-github-token-gcp-secret-name expose the GCP secret name with internal GitHub token as a GitHub Actions organization variable.
# This variable will be available to all repositories in the internal GitHub kyma organization.
# The variable is used in internal github workflows to access the internal GitHub token from GCP Secret Manager.
resource "github_actions_organization_variable" "dora_integration_internal_github_token_gcp_secret_name" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = var.dora_integration_internal_github_token_gcp_secret_name_github_organization_variable
  value         = google_secret_manager_secret.dora_integration_internal_github_token.secret_id
}