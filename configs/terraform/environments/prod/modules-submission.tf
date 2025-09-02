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
