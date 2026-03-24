resource "google_service_account" "kyma-submission-pipeline" {
  account_id   = "kyma-submission-pipeline"
  display_name = "kyma-submission-pipeline"
  #   description = "Service account for retrieving secrets on the conduit-cli build pipeline."
  description = "The submission-pipeline ADO pipeline."
}

module "dev_kyma_modules" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }
  repository_prevent_destroy = var.dev_kyma_modules_repository.repository_prevent_destroy
  repository_name            = var.dev_kyma_modules_repository.name
  description                = var.dev_kyma_modules_repository.description
  repoAdmin_serviceaccounts  = [google_service_account.kyma-submission-pipeline.email]
}

import {
  id = "projects/kyma-project/serviceAccounts/kyma-modules-reader@kyma-project.iam.gserviceaccount.com"
  to = google_service_account.kyma_modules_reader
}

resource "google_service_account" "kyma_modules_reader" {
  provider     = google.kyma_project
  account_id   = "kyma-modules-reader"
  display_name = "kyma-modules-reader"
  description  = "Service account with read-only access to the kyma-modules Artifact Registry."
}

resource "google_artifact_registry_repository_iam_member" "kyma_modules_registry_reader" {
  project    = module.kyma_modules.artifact_registry.project
  repository = module.kyma_modules.artifact_registry.name
  location   = module.kyma_modules.artifact_registry.location
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.kyma_modules_reader.email}"
}

module "kyma_modules" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_prevent_destroy = var.kyma_modules_repository.repository_prevent_destroy
  repository_name            = var.kyma_modules_repository.name
  description                = var.kyma_modules_repository.description
  type                       = var.kyma_modules_repository.type
  reader_serviceaccounts     = var.kyma_modules_repository.reader_serviceaccounts
  reader_groups              = var.kyma_modules_repository.reader_groups
  repoAdmin_serviceaccounts  = [google_service_account.kyma-submission-pipeline.email]
}
