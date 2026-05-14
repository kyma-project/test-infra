data "github_repository" "compliancy_public" {
  provider = github.kyma_project
  name     = "compliancy"
}

data "github_repository" "compliancy_internal" {
  provider = github.internal_github
  name     = "compliancy"
}

resource "google_service_account" "release_log_uploader" {
  account_id  = var.release_log_uploader_service_account_name
  description = "Service account for release log upload workflow in kyma-project/compliancy and kyma/compliancy repositories."
}

resource "google_service_account_iam_member" "release_log_uploader_wif_public_github" {
  service_account_id = google_service_account.release_log_uploader.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.compliancy_public.repo_id}:repository_owner_id:${data.github_organization.kyma_project.id}:workflow:${var.release_log_uploader_workflow_name}"
}

resource "google_service_account_iam_member" "release_log_uploader_wif_internal_github" {
  service_account_id = google_service_account.release_log_uploader.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${local.internal_github_wif_pool_name}/subject/repository_id:${data.github_repository.compliancy_internal.repo_id}:repository_owner_id:${data.github_organization.kyma_internal.id}:workflow:${var.release_log_uploader_workflow_name}"
}

resource "google_storage_bucket_iam_member" "release_log_uploader_logs_bucket_access" {
  bucket = var.release_log_uploader_logs_bucket_name
  role   = "roles/storage.objectUser"
  member = "serviceAccount:${google_service_account.release_log_uploader.email}"
}
