# Release Test Log Storage Configuration
# This bucket stores logs from release tests.
# Used by: https://github.com/kyma-project/compliancy/blob/main/docs/upload-release-report.md

import {
  to = google_storage_bucket.release_test_logs
  id = "kyma_release_test_logs"
}

resource "google_storage_bucket" "release_test_logs" {
  name = "kyma_release_test_logs"
  location = "europe-central2"
  uniform_bucket_level_access = true

  lifecycle {
    prevent_destroy = true
  }
}

# Grant the release-log-uploader service account permission to create AND delete objects in the bucket.
resource "google_storage_bucket_iam_member" "release_log_uploader_access" {
  bucket = google_storage_bucket.release_test_logs.name
  role   = "roles/storage.objectAdmin"

  member = "serviceAccount:release-log-uploader@sap-kyma-prow.iam.gserviceaccount.com"
}

# Grant Kyma developers read-only access to objects in the bucket
resource "google_storage_bucket_iam_member" "kyma_developers_read_access" {
  bucket = google_storage_bucket.release_test_logs.name
  role   = "roles/storage.objectViewer"

  member = "group:kyma_developers@sap.com"
}
