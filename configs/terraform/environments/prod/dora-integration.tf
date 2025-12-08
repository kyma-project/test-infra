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
# GitHub Enterprise (github.tools.sap) instances.
# ==============================================================================

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------

variable "dora-integration-gcp-secret-name-public-github-token" {
  type        = string
  default     = "kyma-bot-github-public-repos-token"
  description = "GCP Secret Manager secret name for public GitHub (github.com) token used by DORA integration"
}

variable "dora-integration-gcp-secret-name-internal-github-token" {
  type        = string
  default     = "kyma-bot-github-internal-repos-token"
  description = "GCP Secret Manager secret name for internal GitHub (github.tools.sap) token used by DORA integration"
}

variable "dora-integration-reusable-workflow-ref" {
  type        = string
  default     = "kyma-project/test-infra/.github/workflows/dora-integration.yml@refs/heads/main"
  description = "Reference to the DORA integration reusable workflow for workload identity federation"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - Public GitHub Token
# ------------------------------------------------------------------------------

# TODO (dekiel): this secret already exists, define resource and import it to manage it via terraform
# Import command: terraform import google_secret_manager_secret.dora-integration-public-github-token projects/sap-kyma-prow/secrets/kyma-bot-github-public-repos-token
resource "google_secret_manager_secret" "dora-integration-public-github-token" {
  project   = var.gcp_project_id
  secret_id = var.dora-integration-gcp-secret-name-public-github-token

  replication {
    auto {}
  }

  labels = {
    type        = "github-token"
    purpose     = "dora-integration"
    github-type = "public"
  }
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - Internal GitHub Token
# ------------------------------------------------------------------------------

# TODO (dekiel): this secret already exists, define resource and import it to manage it via terraform
# Import command: terraform import google_secret_manager_secret.dora-integration-internal-github-token projects/sap-kyma-prow/secrets/kyma-bot-github-internal-repos-token
resource "google_secret_manager_secret" "dora-integration-internal-github-token" {
  project   = var.gcp_project_id
  secret_id = var.dora-integration-gcp-secret-name-internal-github-token

  replication {
    auto {}
  }

  labels = {
    type        = "github-token"
    purpose     = "dora-integration"
    github-type = "internal"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for GitHub Actions Workflows
# ------------------------------------------------------------------------------

# Grant the DORA integration reusable workflow access to read public GitHub token
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_public_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora-integration-public-github-token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.dora-integration-reusable-workflow-ref}"
}

# Grant the DORA integration reusable workflow access to read internal GitHub token
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_internal_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora-integration-internal-github-token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.dora-integration-reusable-workflow-ref}"
}

# ------------------------------------------------------------------------------
# GitHub Actions Organization Variables - Public GitHub (github.com)
# ------------------------------------------------------------------------------

# Expose the public GitHub token secret name as a GitHub Actions organization variable
resource "github_actions_organization_variable" "dora-integration-public-github-token-gcp-secret-name" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "DORA_INTEGRATION_PUBLIC_GITHUB_TOKEN_GCP_SECRET_NAME"
  value         = google_secret_manager_secret.dora-integration-public-github-token.secret_id
}

# ------------------------------------------------------------------------------
# GitHub Actions Organization Variables - Internal GitHub (github.tools.sap)
# ------------------------------------------------------------------------------

# Expose the internal GitHub token secret name as a GitHub Actions organization variable
# This variable will be available in github.tools.sap organization
resource "github_actions_organization_variable" "dora-integration-internal-github-token-gcp-secret-name" {
  provider      = github.github.tools.sap
  visibility    = "all"
  variable_name = "DORA_INTEGRATION_INTERNAL_GITHUB_TOKEN_GCP_SECRET_NAME"
  value         = google_secret_manager_secret.dora-integration-internal-github-token.secret_id
}