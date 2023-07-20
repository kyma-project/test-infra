resource "google_service_account" "sa_gke_kyma_integration" {
  account_id   = "sa_gke_kyma_integration@sap-kyma-prow.iam.gserviceaccount.com"
  display_name = "sa_gke_kyma_integration"
}

resource "google_project_iam_binding" "view_container_analysis_ccurrences" {
  project = var.gcp_project_id
  role    = "roles/containeranalysis.occurrences.viewer"
  members = ["serviceAccount:${google_service_account.sa_gke_kyma_integration.email}"]
}

resource "google_project_iam_binding" "dns_reader" {
  project = var.gcp_project_id
  role    = "roles/dns.reader"
  members = ["serviceAccount:${google_service_account.sa_gke_kyma_integration.email}"]
}

resource "google_project_iam_binding" "bucket_get" {
  project = var.gcp_project_id
  role    = "projects/sap-kyma-prow/roles/BucketGet"
  members = ["serviceAccount:${google_service_account.sa_gke_kyma_integration.email}"]
}