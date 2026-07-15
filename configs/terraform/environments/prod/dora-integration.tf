# ==============================================================================
# DORA Integration Configuration
# ==============================================================================
# This configuration manages resources for DORA (DevOps Research and Assessment)
# metrics collection and reporting in the kyma project.
#
# Resources managed:
# - GCP Secret Manager secrets for GitHub App credentials (App ID and private key)
# - GitHub Actions organization variables for secret references
# - IAM permissions for accessing secrets from GitHub Actions workflows
#
# The configuration supports both public GitHub (github.com) and internal
# GitHub Enterprise instances.
# ==============================================================================

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------

variable "dora_integration_reusable_workflow_ref" {
  type        = string
  default     = "kyma/dora-integration/.github/workflows/dora-integration.yml@refs/heads/main"
  description = "Reference to the DORA integration reusable workflow"
}

variable "dora_integration_gcp_secret_name_githubcom_app_id" {
  type        = string
  default     = "dora-integration-githubcom-app-id"
  description = "GCP Secret Manager secret name for the GitHub App ID used by DORA integration on github.com"
}

variable "dora_integration_gcp_secret_name_githubcom_app_private_key" {
  type        = string
  default     = "dora-integration-githubcom-app-private-key"
  description = "GCP Secret Manager secret name for the GitHub App private key used by DORA integration on github.com"
}

variable "dora_integration_githubcom_app_id_gcp_secret_name_github_organization_variable" {
  type        = string
  default     = "DORA_INTEGRATION_GITHUBCOM_APP_ID_GCP_SECRET_NAME"
  description = "GitHub Actions variable name for the GCP secret name with github.com App ID"
}

variable "dora_integration_githubcom_app_private_key_gcp_secret_name_github_organization_variable" {
  type        = string
  default     = "DORA_INTEGRATION_GITHUBCOM_APP_PRIVATE_KEY_GCP_SECRET_NAME"
  description = "GitHub Actions variable name for the GCP secret name with github.com App private key"
}

variable "dora_integration_gcp_secret_name_githubtoolssap_app_id" {
  type        = string
  default     = "dora-integration-githubtoolssap-app-id"
  description = "GCP Secret Manager secret name for the GitHub App ID used by DORA integration on github.tools.sap"
}

variable "dora_integration_gcp_secret_name_githubtoolssap_app_private_key" {
  type        = string
  default     = "dora-integration-githubtoolssap-app-private-key"
  description = "GCP Secret Manager secret name for the GitHub App private key used by DORA integration on github.tools.sap"
}

variable "dora_integration_githubtoolssap_app_id_gcp_secret_name_github_organization_variable" {
  type        = string
  default     = "DORA_INTEGRATION_GITHUBTOOLSSAP_APP_ID_GCP_SECRET_NAME"
  description = "GitHub Actions variable name for the GCP secret name with github.tools.sap App ID"
}

variable "dora_integration_githubtoolssap_app_private_key_gcp_secret_name_github_organization_variable" {
  type        = string
  default     = "DORA_INTEGRATION_GITHUBTOOLSSAP_APP_PRIVATE_KEY_GCP_SECRET_NAME"
  description = "GitHub Actions variable name for the GCP secret name with github.tools.sap App private key"
}

# ------------------------------------------------------------------------------
# GitHub App Credentials - github.com (kyma-project org)
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret" "dora_integration_githubcom_app_id" {
  project   = var.gcp_project_id
  secret_id = var.dora_integration_gcp_secret_name_githubcom_app_id

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "dora-integration"
    github-instance = "public"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "dora-integration-app"
  }
}

resource "google_secret_manager_secret" "dora_integration_githubcom_app_private_key" {
  project   = var.gcp_project_id
  secret_id = var.dora_integration_gcp_secret_name_githubcom_app_private_key

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "dora-integration"
    github-instance = "public"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "dora-integration-app"
  }
}

# Grant the DORA integration reusable workflow access to read the github.com App ID.
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_githubcom_app_id_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora_integration_githubcom_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.dora_integration_reusable_workflow_ref}"
}

# Grant the DORA integration reusable workflow access to read the github.com App private key.
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_githubcom_app_private_key_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora_integration_githubcom_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.dora_integration_reusable_workflow_ref}"
}

# Expose the GCP secret name with github.com App ID as a GitHub Actions organization variable.
resource "github_actions_organization_variable" "dora_integration_githubcom_app_id_gcp_secret_name" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = var.dora_integration_githubcom_app_id_gcp_secret_name_github_organization_variable
  value         = google_secret_manager_secret.dora_integration_githubcom_app_id.secret_id
}

# Expose the GCP secret name with github.com App private key as a GitHub Actions organization variable.
resource "github_actions_organization_variable" "dora_integration_githubcom_app_private_key_gcp_secret_name" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = var.dora_integration_githubcom_app_private_key_gcp_secret_name_github_organization_variable
  value         = google_secret_manager_secret.dora_integration_githubcom_app_private_key.secret_id
}

# ------------------------------------------------------------------------------
# GitHub App Credentials - github.tools.sap (kyma org)
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret" "dora_integration_githubtoolssap_app_id" {
  project   = var.gcp_project_id
  secret_id = var.dora_integration_gcp_secret_name_githubtoolssap_app_id

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "dora-integration"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "dora-integration-app"
  }
}

resource "google_secret_manager_secret" "dora_integration_githubtoolssap_app_private_key" {
  project   = var.gcp_project_id
  secret_id = var.dora_integration_gcp_secret_name_githubtoolssap_app_private_key

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-credential"
    tool            = "dora-integration"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "reusable-workflow"
    entity          = "dora-integration-app"
  }
}

# Grant the DORA integration reusable workflow access to read the github.tools.sap App ID.
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_githubtoolssap_app_id_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora_integration_githubtoolssap_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.dora_integration_reusable_workflow_ref}"
}

# Grant the DORA integration reusable workflow access to read the github.tools.sap App private key.
resource "google_secret_manager_secret_iam_member" "dora_integration_workflow_githubtoolssap_app_private_key_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.dora_integration_githubtoolssap_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.dora_integration_reusable_workflow_ref}"
}

# Expose the GCP secret name with github.tools.sap App ID as a GitHub Actions organization variable.
resource "github_actions_organization_variable" "dora_integration_githubtoolssap_app_id_gcp_secret_name" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = var.dora_integration_githubtoolssap_app_id_gcp_secret_name_github_organization_variable
  value         = google_secret_manager_secret.dora_integration_githubtoolssap_app_id.secret_id
}

# Expose the GCP secret name with github.tools.sap App private key as a GitHub Actions organization variable.
resource "github_actions_organization_variable" "dora_integration_githubtoolssap_app_private_key_gcp_secret_name" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = var.dora_integration_githubtoolssap_app_private_key_gcp_secret_name_github_organization_variable
  value         = google_secret_manager_secret.dora_integration_githubtoolssap_app_private_key.secret_id
}