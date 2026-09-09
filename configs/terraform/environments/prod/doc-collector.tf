# ==============================================================================
# Documentation Collector Configuration
# ==============================================================================
# This configuration manages resources for the documentation collector workflow
# that synchronizes documentation from various repositories.
#
# Resources managed:
# - GCP Secret Manager secrets for GitHub tokens
# - IAM permissions for accessing secrets via Workload Identity Federation
#
# The workflow runs in kyma/product-kyma-runtime repository and uses
# WIF to authenticate and access secrets from GCP Secret Manager.
# The workflow uses reusable workflow defined in kyma/test-infra repository.
# ==============================================================================

# ------------------------------------------------------------------------------
# Locals
# ------------------------------------------------------------------------------

locals {
  doc_collector_supported_event = [
    "workflow_dispatch",
    "schedule",
    "release",
  ]
}

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------

variable "doc_collector_gcp_secret_name_github_app_private_key" {
  type        = string
  default     = "doc-collector_github-app-private-key"
  description = "GCP Secret Manager secret name for the doc-collector GitHub App private key"
}

variable "doc_collector_gcp_secret_name_github_app_id" {
  type        = string
  default     = "doc-collector_github-app-id"
  description = "GCP Secret Manager secret name for the doc-collector GitHub App ID"
}

variable "doc_collector_reusable_workflow_ref" {
  type = string
  default = "kyma/test-infra/.github/workflows/reusable-doc-collector.yml@refs/heads/main"
  description = "GitHub reference for the reusable workflow used by the documentation collector"
}

# ------------------------------------------------------------------------------
# GitHub Data Sources
# ------------------------------------------------------------------------------

# Fetch the kyma organization data from internal GitHub
data "github_organization" "kyma_internal" {
  provider = github.internal_github
  name     = "kyma"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - GitHub App Private Key
# ------------------------------------------------------------------------------

# Secret shell for the doc-collector GitHub App private key.
# The actual PEM value is uploaded manually after app registration.
resource "google_secret_manager_secret" "doc_collector_github_app_private_key" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_github_app_private_key

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "doc-collector"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "doc-collector-app"
  }
}

# Secret shell for the doc-collector GitHub App ID.
# The value is set manually after app registration.
resource "google_secret_manager_secret" "doc_collector_github_app_id" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_github_app_id

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "doc-collector"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "doc-collector-app"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for GitHub Actions Workflows via WIF
# ------------------------------------------------------------------------------

# Grant the documentation collector workflow access to read the GitHub App private key
# via Workload Identity Federation.
resource "google_secret_manager_secret_iam_member" "doc_collector_reusable_workflow_app_private_key_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_github_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_reusable_workflow_ref}"
}

# Grant the documentation collector workflow access to read the public GitHub token
# (kyma-bot-github-public-repo-token) via Workload Identity Federation.
resource "google_secret_manager_secret_iam_member" "doc_collector_reusable_workflow_public_token_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# Removed Resources - PAT-based authentication (migrated to GitHub App)
# ------------------------------------------------------------------------------

removed {
  from = google_secret_manager_secret.doc_collector_internal_github_token
  lifecycle {
    destroy = false
  }
}

removed {
  from = google_secret_manager_secret_iam_member.doc_collector_reusable_workflow_internal_token_reader
  lifecycle {
    destroy = true
  }
}
