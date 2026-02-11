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

variable "doc_collector_gcp_secret_name_internal_github_token" {
  type        = string
  default     = "technical-writers-docsync-workflow-gh-tools-neighbors-token"
  description = "GCP Secret Manager secret name for internal GitHub token used by documentation collector"
}

variable "doc_collector_workflow_name" {
  type        = string
  default     = "doc-collector"
  description = "Name of the documentation collector workflow"
}

variable "doc_collector_reusable_workflow_ref" {
  type = string
  default = "kyma/test-infra/.github/workflows/reusable-doc-collector.yml@refs/heads/main"
  description = "GitHub reference for the reusable workflow used by the documentation collector"
}

# ------------------------------------------------------------------------------
# GitHub Data Sources
# ------------------------------------------------------------------------------

# Fetch the restricted-markets-docu-hub repository data from internal GitHub
data "github_repository" "restricted_markets_docu_hub" {
  provider = github.internal_github
  name     = "restricted-markets-docu-hub"
}

# Fetch the kyma organization data from internal GitHub
data "github_organization" "kyma_internal" {
  provider = github.internal_github
  name     = "kyma"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - Internal GitHub Token
# ------------------------------------------------------------------------------

# technical-writers-docsync-workflow-gh-tools-neighbors-token stores the Personal Access Token
# for the documentation collector workflow on internal GitHub.
resource "google_secret_manager_secret" "doc_collector_internal_github_token" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_internal_github_token

  replication {
    auto {}
  }

  labels = {
    type            = "github-token"
    tool            = "doc-collector"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "github-workflow"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for GitHub Actions Workflows via WIF
# ------------------------------------------------------------------------------

# Grant the documentation collector workflow access to read the internal GitHub token
# via Workload Identity Federation.
resource "google_secret_manager_secret_iam_member" "doc_collector_workflow_internal_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principal://iam.googleapis.com/${local.internal_github_wif_pool_name}/subject/repository_id:${data.github_repository.restricted_markets_docu_hub.repo_id}:repository_owner_id:${data.github_organization.kyma_internal.id}:workflow:${var.doc_collector_workflow_name}"
}

resource "google_secret_manager_secret_iam_member" "doc_collector_reusable_workflow_internal_token_reader" {
  for_each = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_reusable_workflow_ref}"
}

# Grant the documentation collector workflow access to read the public GitHub token
# (kyma-bot-github-public-repo-token) via Workload Identity Federation.
resource "google_secret_manager_secret_iam_member" "doc_collector_workflow_public_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principal://iam.googleapis.com/${local.internal_github_wif_pool_name}/subject/repository_id:${data.github_repository.restricted_markets_docu_hub.repo_id}:repository_owner_id:${data.github_organization.kyma_internal.id}:workflow:${var.doc_collector_workflow_name}"
}


resource "google_secret_manager_secret_iam_member" "doc_collector_reusable_workflow_public_token_reader" {
  for_each = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_reusable_workflow_ref}"
}

