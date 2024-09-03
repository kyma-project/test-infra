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


# Grant owner role to terraform executor service account in the gcp workloads project.
resource "google_project_iam_member" "terraform_executor_workloads_project_owner" {
  project = var.workloads_project_id
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.terraform_executor.email}"
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
    "roles/iam.securityReviewer",
    "roles/container.developer" # TODO(KacperMalachowski): Remove when Prow will be removed
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
    "principal://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/subject/repository_id:147495537:repository_owner_id:39153523:workflow:Pull Plan Prod Terraform"
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.terraform_planner.name
}


resource "google_project_iam_member" "terraform_planner_workloads_project_read_access" {
  for_each = toset([
    "roles/viewer",
  ])
  project = var.workloads_project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "google_service_account_iam_member" "terraform_executor_workload_identity_user" {
  member = "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_apply_workflow_name}"
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_executor.email}"
}

resource "google_service_account_iam_member" "terraform_planner_workload_identity_user" {
  member = "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_plan_workflow_name}"
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