#Service account and IAM bindings are needed for development/tools/cmd/dnscollector app which is used on orphaned-dns-cleaner prow job.

resource "google_service_account" "sa_gke_kyma_integration" {
  account_id   = "sa-gke-kyma-integration"
  display_name = "sa-gke-kyma-integration"
  description  = "Service account is used by Prow to integrate with GKE. Will be removed with Prow"
}

resource "google_project_iam_binding" "dns_collector_container_analysis_occurrences_viewer" {
  project = var.gcp_project_id
  role    = "roles/containeranalysis.occurrences.viewer"
  members = ["serviceAccount:${google_service_account.sa_gke_kyma_integration.email}"]
}

resource "google_project_iam_binding" "dns_collector_dns_reader" {
  project = var.gcp_project_id
  role    = "roles/dns.reader"
  members = ["serviceAccount:${google_service_account.sa_gke_kyma_integration.email}"]
}

resource "google_project_iam_binding" "dns_collector_bucket_get" {
  project = var.gcp_project_id
  role    = "projects/sap-kyma-prow/roles/BucketGet"
  members = ["serviceAccount:${google_service_account.sa_gke_kyma_integration.email}"]
}