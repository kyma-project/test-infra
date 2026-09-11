# ==============================================================================
# Documentation Collector Configuration
# ==============================================================================
# This configuration manages resources for the documentation collector workflow
# that synchronizes documentation from various repositories.
#
# Resources managed:
# - GCP Secret Manager secrets for GitHub App credentials
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

variable "doc_collector_gcp_secret_name_internal_app_private_key" {
  type        = string
  default     = "doc-collector_internal-github-app-private-key"
  description = "GCP Secret Manager secret name for the doc-collector github.tools.sap App private key"
}

variable "doc_collector_gcp_secret_name_internal_app_id" {
  type        = string
  default     = "doc-collector_internal-github-app-id"
  description = "GCP Secret Manager secret name for the doc-collector github.tools.sap App ID"
}

variable "doc_collector_gcp_secret_name_public_app_private_key" {
  type        = string
  default     = "doc-collector_public-github-app-private-key"
  description = "GCP Secret Manager secret name for the doc-collector github.com App private key"
}

variable "doc_collector_gcp_secret_name_public_app_id" {
  type        = string
  default     = "doc-collector_public-github-app-id"
  description = "GCP Secret Manager secret name for the doc-collector github.com App ID"
}

variable "doc_collector_internal_reusable_workflow_ref" {
  type        = string
  default     = "kyma/test-infra/.github/workflows/reusable-doc-collector.yml@refs/heads/main"
  description = "GitHub reference for the reusable workflow on github.tools.sap used by the documentation collector"
}

# ------------------------------------------------------------------------------
# GitHub Data Sources
# ------------------------------------------------------------------------------

data "github_organization" "kyma_internal" {
  provider = github.internal_github
  name     = "kyma"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager - github.tools.sap App credentials
# ------------------------------------------------------------------------------

# Secret shell for the doc-collector github.tools.sap App private key.
# The actual PEM value is uploaded manually after app registration.
resource "google_secret_manager_secret" "doc_collector_internal_app_private_key" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_internal_app_private_key

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

# Secret shell for the doc-collector github.tools.sap App ID.
# The value is set manually after app registration.
resource "google_secret_manager_secret" "doc_collector_internal_app_id" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_internal_app_id

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
# GCP Secret Manager - github.com App credentials
# ------------------------------------------------------------------------------

# Secret shell for the doc-collector github.com App private key.
# The actual PEM value is uploaded manually after app registration.
resource "google_secret_manager_secret" "doc_collector_public_app_private_key" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_public_app_private_key

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "doc-collector"
    github-instance = "public"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "doc-collector-app"
  }
}

# Secret shell for the doc-collector github.com App ID.
# The value is set manually after app registration.
resource "google_secret_manager_secret" "doc_collector_public_app_id" {
  project   = var.gcp_project_id
  secret_id = var.doc_collector_gcp_secret_name_public_app_id

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "doc-collector"
    github-instance = "public"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "doc-collector-app"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - github.tools.sap App
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "doc_collector_internal_app_private_key_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_internal_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_internal_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "doc_collector_internal_app_id_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_internal_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_internal_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# IAM Permissions - github.com App
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "doc_collector_public_app_private_key_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_public_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_internal_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "doc_collector_public_app_id_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_public_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_internal_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# TODO: remove after GitHub App auth is validated in production
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "doc_collector_reusable_workflow_public_token_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_internal_reusable_workflow_ref}"
}

variable "doc_collector_gcp_secret_name_internal_github_token" {
  type        = string
  default     = "technical-writers-docsync-workflow-gh-tools-neighbors-token"
  description = "GCP Secret Manager secret name for internal GitHub token used by documentation collector"
}

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

resource "google_secret_manager_secret_iam_member" "doc_collector_reusable_workflow_internal_token_reader" {
  for_each  = toset(local.doc_collector_supported_event)
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.doc_collector_internal_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${each.value}:repository_owner_id:${data.github_organization.kyma_internal.id}:reusable_workflow_ref:${var.doc_collector_internal_reusable_workflow_ref}"
}
