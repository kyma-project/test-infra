resource "google_service_account" "sa_gke_kyma_integration" {
  account_id   = "sa-gke-kyma-integration"
  display_name = "sa-gke-kyma-integration"
  description  = "Service account is used by Prow to integrate with GKE."
}

resource "google_project_iam_binding" "dns_cleaner_view_container_analysis_occurrences" {
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