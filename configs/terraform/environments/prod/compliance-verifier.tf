resource "google_service_account" "kyma-compliance-pipeline" {
  account_id   = "kyma-compliance-pipeline"
  display_name = "kyma-compliance-pipeline"
  description  = "Service account for retrieving secrets on the compliance Azure pipeline."

  lifecycle {
    prevent_destroy = true
  }
}

# Grant read access to the sec-scanner-cfg-gcp-sa-key secret for kyma-compliance-pipeline service account.
resource "google_secret_manager_secret_iam_member" "compliance_verifier_sec_scanner_cfg_secret_accessor" {
  project   = data.google_client_config.gcp.project
  secret_id = google_secret_manager_secret.sec-scanner-cfg-processor-gcp-sa-key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kyma-compliance-pipeline.email}"
}
