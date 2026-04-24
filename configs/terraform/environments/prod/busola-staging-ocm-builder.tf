resource "google_service_account" "busola_staging_ocm_builder" {
  account_id   = "busola-staging-ocm-builder"
  display_name = "busola-staging-ocm-builder"
  description  = "Service account for building Busola Staging OCM components."
}

removed {
  from = google_artifact_registry_repository_iam_member.busola_staging_ocm_builder_dev_kyma_modules_writer
  lifecycle { destroy = false }
}

# Grant read access to restricted-prod Artifact Registry
resource "google_artifact_registry_repository_iam_member" "busola_staging_ocm_builder_restricted_prod_reader" {
  provider   = google.kyma_project
  location   = module.restricted_prod.artifact_registry.location
  repository = module.restricted_prod.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.busola_staging_ocm_builder.email}"
}

# Grant read access to prod Artifact Registry
resource "google_artifact_registry_repository_iam_member" "busola_staging_ocm_builder_prod_reader" {
  provider   = google.kyma_project
  location   = module.prod_docker_repository.artifact_registry.location
  repository = module.prod_docker_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.busola_staging_ocm_builder.email}"
}
