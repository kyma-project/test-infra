data "github_repository" "tooling_infra" {
  provider = github.internal_github
  name     = "tooling-infra"
}

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

# Grant workflows the workload identity user role on the terraform executor service account.
# This is required to let workflows impersonate the terraform executor service account.
# Authentication is done through OIDC providers and Google Workload Identity Federation.
#
# IMPORTANT: This is an authoritative binding for roles/iam.workloadIdentityUser.
# All principals that need this role MUST be listed here. Do NOT use
# google_service_account_iam_member for the same role — it will be overwritten.
resource "google_service_account_iam_binding" "terraform_workload_identity" {
  members = [
    # github.com (kyma-project) — post-apply-prod-terraform workflow
    "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_apply_workflow_name}",

    # Internal GitHub Enterprise (github-tools-sap) — tooling-infra deploy workflow (prod, v-tag)
    "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.deploy_identity/${var.internal_github_tooling_infra_terraform_deploy_identity_prod}",

    # Internal GitHub Enterprise (github-tools-sap) — tooling-infra deploy workflow (staging branch)
    "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.deploy_identity/${var.internal_github_tooling_infra_terraform_deploy_identity_staging}",
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

# Grant workflows the workload identity user role on the terraform planner service account.
#
# IMPORTANT: This is an authoritative binding for roles/iam.workloadIdentityUser.
# All principals that need this role MUST be listed here. Do NOT use
# google_service_account_iam_member for the same role — it will be overwritten.
resource "google_service_account_iam_binding" "terraform_planner_workload_identity" {
  members = [
    # github.com (kyma-project) — direct workflow triggers
    "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_plan_workflow_name}",
    "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:Tofu Drift Detection",

    # github.com (kyma-project) — reusable workflow triggers for plan prod terraform
    "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:merge_group:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-plan-prod-terraform.yaml@refs/heads/main",
    "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:pull_request_target:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-plan-prod-terraform.yaml@refs/heads/main",

    # github.com (kyma-project) — reusable workflow triggers for validate service accounts
    "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:pull_request_target:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-validate-service-accounts.yaml@refs/heads/main",
    "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:merge_group:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:kyma-project/test-infra/.github/workflows/pull-validate-service-accounts.yaml@refs/heads/main",

    # Internal GitHub Enterprise (github-tools-sap) — tooling-infra plan workflow
    "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.internal_github_tooling_infra_terraform_plan_reusable_workflow_ref}",

    # Internal GitHub Enterprise (github-tools-sap) — tooling-infra validate workflow
    "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.internal_github_tooling_infra_terraform_validate_reusable_workflow_ref}",

    # Internal GitHub Enterprise (github-tools-sap) — any workflow in kyma/tooling-infra
    "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.repository_id/${data.github_repository.tooling_infra.repo_id}",
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.terraform_planner.name
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

# GitHub App credentials for kyma-project-terraform-planner
resource "github_actions_variable" "terraform_planner_github_app_id_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_PLANNER_APP_ID_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_planner_github_app_id.secret_id
}

resource "github_actions_variable" "terraform_planner_github_app_private_key_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_PLANNER_APP_PRIVATE_KEY_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_planner_github_app_private_key.secret_id
}

# GitHub App credentials for kyma-project-terraform-executor
resource "github_actions_variable" "terraform_executor_github_app_id_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_EXECUTOR_APP_ID_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_executor_github_app_id.secret_id
}

resource "github_actions_variable" "terraform_executor_github_app_private_key_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_EXECUTOR_APP_PRIVATE_KEY_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_executor_github_app_private_key.secret_id
}

# ==============================================================================
# GitHub App Authentication - kyma-project-terraform-executor
# ==============================================================================
# The app has write permissions for org/repo
# variables and organization administration (hosted runners).
# Secret values must be added manually via GCP Console or CLI after creation.
# ==============================================================================

resource "google_secret_manager_secret" "terraform_executor_github_app_id" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "kyma-project-terraform-executor-app-id"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-id"
    tool            = "iac"
    github-instance = "public"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret" "terraform_executor_github_app_private_key" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "kyma-project-terraform-executor-app-private-key"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-private-key"
    tool            = "iac"
    github-instance = "public"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret_iam_member" "terraform_executor_github_app_id_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_executor_github_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_executor.email}"
}

resource "google_secret_manager_secret_iam_member" "terraform_executor_github_app_private_key_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_executor_github_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_executor.email}"
}

# ==============================================================================
# GitHub App Authentication - kyma-project-terraform-planner
# ==============================================================================
# The app has read-only permissions for org/repo
# variables and organization administration (hosted runners).
# Secret values must be added manually via GCP Console or CLI after creation.
# ==============================================================================

resource "google_secret_manager_secret" "terraform_planner_github_app_id" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "kyma-project-terraform-planner-app-id"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-id"
    tool            = "iac"
    github-instance = "public"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret" "terraform_planner_github_app_private_key" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "kyma-project-terraform-planner-app-private-key"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-private-key"
    tool            = "iac"
    github-instance = "public"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret_iam_member" "terraform_planner_github_app_id_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_planner_github_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "google_secret_manager_secret_iam_member" "terraform_planner_github_app_private_key_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_planner_github_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

# ==============================================================================
# GitHub App Authentication — internal GitHub (github.tools.sap) — executor
# ==============================================================================
# Replaces the long-lived PAT (iac-bot-gh-tools-sap-terraform-executor-token).
# ==============================================================================

resource "google_secret_manager_secret" "terraform_executor_internal_github_app_id" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "terraform-executor_internal-github-app-id"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-id"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret" "terraform_executor_internal_github_app_private_key" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "terraform-executor_internal-github-app-private-key"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-private-key"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret_iam_member" "terraform_executor_internal_github_app_id_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_executor_internal_github_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_executor.email}"
}

resource "google_secret_manager_secret_iam_member" "terraform_executor_internal_github_app_private_key_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_executor_internal_github_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_executor.email}"
}

resource "github_actions_variable" "terraform_executor_internal_github_app_id_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_EXECUTOR_INTERNAL_APP_ID_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_executor_internal_github_app_id.secret_id
}

resource "github_actions_variable" "terraform_executor_internal_github_app_private_key_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_EXECUTOR_INTERNAL_APP_PRIVATE_KEY_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_executor_internal_github_app_private_key.secret_id
}

# ==============================================================================
# GitHub App Authentication — internal GitHub (github.tools.sap) — planner
# ==============================================================================
# Replaces the long-lived PAT (iac-bot-gh-tools-sap-terraform-planner-token).
# Secret values must be added manually after App creation.
# ==============================================================================

resource "google_secret_manager_secret" "terraform_planner_internal_github_app_id" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "terraform-planner_internal-github-app-id"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-id"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret" "terraform_planner_internal_github_app_private_key" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "terraform-planner_internal-github-app-private-key"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-private-key"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret_iam_member" "terraform_planner_internal_github_app_id_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_planner_internal_github_app_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "google_secret_manager_secret_iam_member" "terraform_planner_internal_github_app_private_key_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_planner_internal_github_app_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "github_actions_variable" "terraform_planner_internal_github_app_id_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_PLANNER_INTERNAL_APP_ID_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_planner_internal_github_app_id.secret_id
}

resource "github_actions_variable" "terraform_planner_internal_github_app_private_key_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_PLANNER_INTERNAL_APP_PRIVATE_KEY_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_planner_internal_github_app_private_key.secret_id
}

# ==============================================================================
# GitHub App Authentication — internal GitHub (github.tools.sap) — installation IDs
# ==============================================================================
# Installation ID is not sensitive but stored as a secret for consistency.
# ==============================================================================

resource "google_secret_manager_secret" "terraform_executor_internal_github_app_installation_id" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "terraform-executor_internal-github-app-installation-id"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-installation-id"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret_iam_member" "terraform_executor_internal_github_app_installation_id_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_executor_internal_github_app_installation_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_executor.email}"
}

resource "github_actions_variable" "terraform_executor_internal_github_app_installation_id_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_EXECUTOR_INTERNAL_APP_INSTALLATION_ID_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_executor_internal_github_app_installation_id.secret_id
}

resource "google_secret_manager_secret" "terraform_planner_internal_github_app_installation_id" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = "terraform-planner_internal-github-app-installation-id"

  replication {
    auto {}
  }

  labels = {
    type            = "github-app-installation-id"
    tool            = "iac"
    github-instance = "internal"
    owner           = "neighbors"
  }
}

resource "google_secret_manager_secret_iam_member" "terraform_planner_internal_github_app_installation_id_reader" {
  project   = var.terraform_executor_gcp_service_account.project_id
  secret_id = google_secret_manager_secret.terraform_planner_internal_github_app_installation_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "github_actions_variable" "terraform_planner_internal_github_app_installation_id_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_TERRAFORM_PLANNER_INTERNAL_APP_INSTALLATION_ID_SECRET_NAME"
  value         = google_secret_manager_secret.terraform_planner_internal_github_app_installation_id.secret_id
}
