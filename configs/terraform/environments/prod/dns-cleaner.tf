resource "google_service_account" "sa-gke-kyma-integration" {
  account_id   = "sa-gke-kyma-integration@sap-kyma-prow.iam.gserviceaccount.com"
  display_name = "sa-gke-kyma-integration"
}

resource "google_project_iam_binding" "view-container-analysis-ccurrences" {
  project = var.gcp_project_id
  role    = "roles/containeranalysis.occurrences.viewer"
  members = ["serviceAccount:${google_service_account.sa-gke-kyma-integration.email}"]
}

resource "google_project_iam_binding" "dns-reader" {
  project = var.gcp_project_id
  role    = "roles/dns.reader"
  members = ["serviceAccount:${google_service_account.sa-gke-kyma-integration.email}"]
}

resource "google_project_iam_binding" "bucket-get" {
  project = "sap-kyma-prow"
  role    = "projects/sap-kyma-prow/roles/BucketGet"
  members = ["serviceAccount:${google_service_account.sa-gke-kyma-integration.email}"]
}