# Release Test Logs
# Manages GCS bucket and service account for storing release test logs.
# Tool source: https://github.com/kyma-project/compliancy/blob/main/docs/upload-release-report.md

# --------------------------------------------------------------------------
# Variables
# --------------------------------------------------------------------------

variable "release_report_workflow_name" {
  type        = string
  default     = "Release report"
  description = "Name of the release report workflow in kyma/compliancy that uploads logs"
}

# --------------------------------------------------------------------------
# Data Sources
# --------------------------------------------------------------------------

data "github_repository" "compliancy" {
  provider = github.internal_github
  name     = "compliancy"
}

# --------------------------------------------------------------------------
# GCS Bucket
# --------------------------------------------------------------------------

import {
  to = google_storage_bucket.release_test_logs
  id = "sap-kyma-prow/kyma_release_test_logs"
}

resource "google_storage_bucket" "release_test_logs" {
  name                        = "kyma_release_test_logs"
  project                     = "sap-kyma-prow"
  location                    = "EUROPE-CENTRAL2"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true

  lifecycle {
    prevent_destroy = true
  }
}

# --------------------------------------------------------------------------
# Service Account
# --------------------------------------------------------------------------

import {
  to = google_service_account.release_log_uploader
  id = "projects/sap-kyma-prow/serviceAccounts/release-log-uploader@sap-kyma-prow.iam.gserviceaccount.com"
}

resource "google_service_account" "release_log_uploader" {
  project      = "sap-kyma-prow"
  account_id   = "release-log-uploader"
  display_name = "release-log-uploader"
  description  = "Uploads release test logs to GCS bucket kyma_release_test_logs"
}

# --------------------------------------------------------------------------
# Workload Identity Federation - allow GitHub Actions to authenticate as SA
# --------------------------------------------------------------------------

resource "google_service_account_iam_member" "release_log_uploader_wif_internal" {
  service_account_id = google_service_account.release_log_uploader.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${local.internal_github_wif_pool_name}/subject/repository_id:${data.github_repository.compliancy.repo_id}:repository_owner_id:${data.github_organization.kyma_internal.id}:workflow:${var.release_report_workflow_name}"
}

# --------------------------------------------------------------------------
# Bucket IAM
# --------------------------------------------------------------------------

resource "google_storage_bucket_iam_member" "release_log_uploader_access" {
  bucket = google_storage_bucket.release_test_logs.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.release_log_uploader.email}"
}

resource "google_storage_bucket_iam_member" "kyma_developers_read_access" {
  bucket = google_storage_bucket.release_test_logs.name
  role   = "roles/storage.objectViewer"
  member = "group:kyma_developers@sap.com"
}
