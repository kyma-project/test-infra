# ==============================================================================
# Go CI Reusable Workflows - GitHub App Access
# ==============================================================================

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

# ==============================================================================
# test-infra-private-repo-access — GitHub App credentials + WIF access
# ==============================================================================

variable "test_infra_private_repo_access_internal_app_id_secret_id" {
  type        = string
  default     = "test-infra-private-repo-access_internal-github-app-id"
  description = "GCP Secret Manager secret ID for the test-infra-private-repo-access GitHub App ID on the internal GitHub (github.tools.sap)"
}

variable "test_infra_private_repo_access_internal_private_key_secret_id" {
  type        = string
  default     = "test-infra-private-repo-access_internal-github-app-private-key"
  description = "GCP Secret Manager secret ID for the test-infra-private-repo-access GitHub App private key on the internal GitHub (github.tools.sap)"
}

variable "gh_test_infra_private_repo_access_internal_app_id_secret_var_name" {
  type        = string
  default     = "GH_TEST_INFRA_PRIVATE_REPO_ACCESS_INTERNAL_APP_ID_SECRET_NAME"
  description = "GitHub Actions repository variable name that holds the GCP secret name for the test-infra-private-repo-access App ID"
}

variable "gh_test_infra_private_repo_access_internal_private_key_secret_var_name" {
  type        = string
  default     = "GH_TEST_INFRA_PRIVATE_REPO_ACCESS_INTERNAL_PRIVATE_KEY_SECRET_NAME"
  description = "GitHub Actions repository variable name that holds the GCP secret name for the test-infra-private-repo-access App private key"
}

# ------------------------------------------------------------------------------
# GCP Secret Manager — secret containers
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret" "test_infra_private_repo_access_internal_app_id" {
  project   = var.gcp_project_id
  secret_id = var.test_infra_private_repo_access_internal_app_id_secret_id

  replication {
    auto {}
  }

  labels = {
    owner           = "neighbors"
    type            = "app-id"
    tool            = "test-infra-private-repo-access"
    entity          = "github-app"
    github-instance = "internal"
    managed-by      = "terraform"
  }
}

resource "google_secret_manager_secret" "test_infra_private_repo_access_internal_private_key" {
  project   = var.gcp_project_id
  secret_id = var.test_infra_private_repo_access_internal_private_key_secret_id

  replication {
    auto {}
  }

  labels = {
    owner           = "neighbors"
    type            = "private-key"
    tool            = "test-infra-private-repo-access"
    entity          = "github-app"
    github-instance = "internal"
    managed-by      = "terraform"
  }
}

# ------------------------------------------------------------------------------
# IAM — grant pull-go-lint reusable workflow secretAccessor via WIF
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "pull_go_lint_test_infra_private_repo_access_internal_app_id_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.test_infra_private_repo_access_internal_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_go_lint_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "pull_go_lint_test_infra_private_repo_access_internal_private_key_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.test_infra_private_repo_access_internal_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_go_lint_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# IAM — grant pull-unit-test-go reusable workflow secretAccessor via WIF
# ------------------------------------------------------------------------------

resource "google_secret_manager_secret_iam_member" "pull_unit_test_go_test_infra_private_repo_access_internal_app_id_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.test_infra_private_repo_access_internal_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_unit_test_go_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "pull_unit_test_go_test_infra_private_repo_access_internal_private_key_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.test_infra_private_repo_access_internal_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.pull_unit_test_go_reusable_workflow_ref}"
}

# ------------------------------------------------------------------------------
# GitHub Actions Repository Variables — expose secret names to reusable workflows
# ------------------------------------------------------------------------------

resource "github_actions_variable" "test_infra_private_repo_access_internal_app_id_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = var.gh_test_infra_private_repo_access_internal_app_id_secret_var_name
  value         = google_secret_manager_secret.test_infra_private_repo_access_internal_app_id.secret_id
}

resource "github_actions_variable" "test_infra_private_repo_access_internal_private_key_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = var.gh_test_infra_private_repo_access_internal_private_key_secret_var_name
  value         = google_secret_manager_secret.test_infra_private_repo_access_internal_private_key.secret_id
}
