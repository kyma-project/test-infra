resource "google_service_account" "restricted-markets-artifactregistry-reader" {
  account_id   = var.sre-restricted-markets-artifactregistry-reader.registry-reader-sa
  display_name = var.sre-restricted-markets-artifactregistry-reader.registry-reader-sa
  description  = var.sre-restricted-markets-artifactregistry-reader.sa-description
}

resource "google_service_account_iam_member" "restricted_markets_artifactregistry_reader_impersonation" {
  service_account_id = google_service_account.restricted-markets-artifactregistry-reader.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${var.sre-restricted-markets-artifactregistry-reader.sre-registry-reader-sa}"
}

resource "google_artifact_registry_repository_iam_member" "kyma_modules_reader" {
  provider   = google.kyma_project
  repository = module.kyma_modules.artifact_registry_collection.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.restricted-markets-artifactregistry-reader.email}"
}

variable "sre-restricted-markets-artifactregistry-reader" {
  type = object({
    registry-reader-sa     = string
    sre-registry-reader-sa = string
  })
  default = {
    registry-reader-sa     = "restricted-markets-reg-reader"
    sa-description         = "Service account for restricted markets delivery with artifact registry reader access"
    sre-registry-reader-sa = "gcr-writer@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com"
  }
}