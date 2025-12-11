# Create the terraform executor Google Cloud service account.
# It grants owner rights to the Google Cloud service account. The owner role is required to let
# the terraform executor manage all the resources in the Google Cloud project.
# It also grants the terraform executor gcp service account the owner role in the workloads project.

resource "google_service_account" "terraform_executor" {
  project      = var.terraform_executor_gcp_service_account.project_id
  account_id   = var.terraform_executor_gcp_service_account.id
  display_name = var.terraform_executor_gcp_service_account.id
  description  = "Identity of terraform executor. It's mapped to k8s service account through workload identity."
}

# Grant owner role to terraform executor service account in the prow project.
resource "google_project_iam_member" "terraform_executor_prow_project_owner" {
  project = var.terraform_executor_gcp_service_account.project_id
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.terraform_executor.email}"
}

# Grant pull-plan-prod-terraform and post-apply-prod-terraform workflows the workload identity user role in the terraform executor service account.
# This is required to let the workflow impersonate the terraform executor service account.
# Authentication is done through github oidc provider and google workload identity federation.
resource "google_service_account_iam_binding" "terraform_workload_identity" {
  members = [
    "principal://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/subject/repository_id:147495537:repository_owner_id:39153523:workflow:Post Apply Prod Terraform"
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.terraform_executor.name
}

# Create the terraform planner GCP service account.
# Grants the browser permissions to refresh state of the resources.

resource "google_service_account" "terraform_planner" {
  project      = var.terraform_planner_gcp_service_account.project_id
  account_id   = var.terraform_planner_gcp_service_account.id
  display_name = var.terraform_planner_gcp_service_account.id

  description = "Identity of terraform planner"
}

# Grant viewer role to terraform planner service account
resource "google_project_iam_member" "terraform_planner_prow_project_read_access" {
  for_each = toset([
    "roles/viewer",
    "roles/storage.objectViewer",
    "roles/iam.securityReviewer"
  ])
  project = var.terraform_planner_gcp_service_account.project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "google_storage_bucket_iam_binding" "planner_state_bucket_write_access" {
  bucket = "tf-state-kyma-project"
  members = [
    "serviceAccount:${google_service_account.terraform_planner.email}"
  ]
  role = "roles/storage.objectUser"
}

resource "google_service_account_iam_binding" "terraform_planner_workload_identity" {
  members = [
    "principal://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/subject/repository_id:147495537:repository_owner_id:39153523:workflow:Pull Plan Prod Terraform",

    # This is used by the reusable workflow to run the plan prod terraform workflow
    "principalSet://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/attribute.reusable_workflow_run/event_name:merge_group:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-plan-prod-terraform.yaml@refs/heads/main",
    "principalSet://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/attribute.reusable_workflow_run/event_name:pull_request_target:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-plan-prod-terraform.yaml@refs/heads/main",

    # This is used by the reusable workflow to run the pull-validate-service-accounts workflow
    "principalSet://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/attribute.reusable_workflow_run/event_name:pull_request_target:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-validate-service-accounts.yaml@refs/heads/main",
    "principalSet://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/attribute.reusable_workflow_run/event_name:merge_group:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-validate-service-accounts.yaml@refs/heads/main"
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.terraform_planner.name
}

resource "google_service_account_iam_member" "terraform_executor_workload_identity_user" {
  member             = "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_apply_workflow_name}"
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_executor.email}"
}

resource "google_service_account_iam_member" "terraform_planner_workload_identity_user" {
  member             = "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_plan_workflow_name}"
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_planner.email}"
}

resource "github_actions_variable" "gcp_terraform_executor_service_account_email" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GCP_TERRAFORM_EXECUTOR_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.terraform_executor.email
}

resource "github_actions_variable" "gcp_terraform_planner_service_account_email" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GCP_TERRAFORM_PLANNER_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.terraform_planner.email
}

# ------------------------------------------------------------------------------
# GitHub Actions Variables for github.com Token Secret Names
# ------------------------------------------------------------------------------
# These variables expose the GCP Secret Manager secret names to GitHub Actions
# workflows. Workflows use these variable names to retrieve the actual tokens
# from GCP Secret Manager during execution.
# ------------------------------------------------------------------------------

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

# ------------------------------------------------------------------------------
# Internal GitHub Token Configuration
# ------------------------------------------------------------------------------
# This section manages GCP Secret Manager secrets for internal GitHub tokens.
# These tokens are used by Terraform workflows to authenticate to SAP's internal
# GitHub Enterprise instance when managing resources like organization variables.
#
# Two separate tokens are maintained:
# - Executor token: Used by post-apply-prod-terraform.yaml for write operations
# - Planner token: Used by pull-plan-prod-terraform.yaml for read-only operations
#
# The secret names are exposed as GitHub Actions variables so workflows can
# retrieve the tokens from GCP Secret Manager during execution.
# ------------------------------------------------------------------------------

variable "internal_github_terraform_executor_secret_name" {
  description = "Name of the GCP secret manager's secret holding iac service user token for internal GitHub terraform executor"
  type        = string
  default     = "iac-bot-gh-tools-sap-terraform-executor-token"
}

variable "internal_github_terraform_planner_secret_name" {
  description = "Name of the GCP secret manager's secret holding iac service user token for internal GitHub terraform planner"
  type        = string
  default     = "iac-bot-gh-tools-sap-terraform-planner-token"
}

variable "internal_github_terraform_executor_variable_name" {
  description = "Name of the GitHub Actions variable for GCP secret name holding internal GitHub terraform executor token"
  type        = string
  default     = "INTERNAL_GITHUB_TERRAFORM_EXECUTOR_SECRET_NAME"
}

variable "internal_github_terraform_planner_variable_name" {
  description = "Name of the GitHub Actions variable for GCP secret name holding internal GitHub terraform planner token"
  type        = string
  default     = "INTERNAL_GITHUB_TERRAFORM_PLANNER_SECRET_NAME"
}

import {
  to = google_secret_manager_secret.internal_github_terraform_executor
  id = "projects/${var.terraform_executor_gcp_service_account.project_id}/secrets/${var.internal_github_terraform_executor_secret_name}"
}

# GCP Secret Manager secret for internal GitHub terraform executor token.
# This token should have write permissions and is used during terraform apply.
# The actual token value must be added manually via GCP Console or CLI.
resource "google_secret_manager_secret" "internal_github_terraform_executor" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = var.internal_github_terraform_executor_secret_name

  replication {
    auto {}
  }

  labels = {
    type            = "github-token"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

import {
  to = google_secret_manager_secret.internal_github_terraform_planner
  id = "projects/${var.terraform_executor_gcp_service_account.project_id}/secrets/${var.internal_github_terraform_planner_secret_name}"
}

# GCP Secret Manager secret for internal GitHub terraform planner token.
# This token should have read-only permissions and is used during terraform plan.
# The actual token value must be added manually via GCP Console or CLI.
resource "google_secret_manager_secret" "internal_github_terraform_planner" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = var.internal_github_terraform_planner_secret_name

  replication {
    auto {}
  }

  labels = {
    type            = "github-token"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

# ------------------------------------------------------------------------------
# IAM Permissions - Secret Access for Terraform Planner
# ------------------------------------------------------------------------------
# Grant the terraform planner service account access to read internal GitHub tokens.
# This is required for terraform plan operations that need to read the current
# state of internal GitHub resources.
# ------------------------------------------------------------------------------

# terraform_planner_internal_github_executor_secret_reader grants the terraform planner
# service account read access to the internal GitHub executor token secret.
# IMPORTANT: This should be part of bootstrapp process to let terraform planner read the secret.
resource "google_secret_manager_secret_iam_member" "terraform_planner_internal_github_executor_secret_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.internal_github_terraform_executor.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

# terraform_planner_internal_github_planner_secret_reader grants the terraform planner
# service account read access to the internal GitHub planner token secret.
# IMPORTANT: This should be part of bootstrapp process to let terraform planner read the secret.
resource "google_secret_manager_secret_iam_member" "terraform_planner_internal_github_planner_secret_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.internal_github_terraform_planner.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

import {
  to = github_actions_variable.internal_github_terraform_executor_secret_name
  id = "test-infra:${var.internal_github_terraform_executor_variable_name}"
}

# internal_github_terraform_executor_variable_name exposes the GCP secret name with internal GitHub token for IaC executor as a GitHub Actions repository variable.
# The variable has repository scope.
# IMPORTANT: This should be part of bootstrapp process to let terraform executor read the secret.
resource "github_actions_variable" "internal_github_terraform_executor_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = var.internal_github_terraform_executor_variable_name
  value         = google_secret_manager_secret.internal_github_terraform_executor.secret_id
}

import {
  to = github_actions_variable.internal_github_terraform_planner_secret_name
  id = "test-infra:${var.internal_github_terraform_planner_variable_name}"
}

# internal_github_terraform_planner_secret_name exposes the GCP secret name with internal GitHub token for IaC planner as a GitHub Actions repository variable.
# The variable has repository scope.
# IMPORTANT: This should be part of bootstrapp process to let terraform planner read the secret.
resource "github_actions_variable" "internal_github_terraform_planner_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = var.internal_github_terraform_planner_variable_name
  value         = google_secret_manager_secret.internal_github_terraform_planner.secret_id
}