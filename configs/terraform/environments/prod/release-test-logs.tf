# ==============================================================================
# Release Test Logs Bucket
# ==============================================================================
# Imports the existing kyma_release_test_logs GCS bucket into Terraform and
# grants Kyma developers read access.
#
# The service account, WIF bindings, and SA bucket permissions are managed in
# release-log-uploader.tf.
#
# Tool source: https://github.com/kyma-project/compliancy/blob/main/docs/upload-release-report.md
# ==============================================================================

# GCS Bucket

import {
  to = google_storage_bucket.release_test_logs
  id = "sap-kyma-prow/kyma_release_test_logs"
}

resource "google_storage_bucket" "release_test_logs" {
  name                        = "kyma_release_test_logs"
  project                     = "sap-kyma-prow"
  location                    = "EU"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true

  lifecycle {
    prevent_destroy = true
  }
}

# Bucket IAM - developer read access

resource "google_storage_bucket_iam_member" "kyma_developers_read_access" {
  bucket = google_storage_bucket.release_test_logs.name
  role   = "roles/storage.objectViewer"
  member = "group:kyma_developers@sap.com"
}
