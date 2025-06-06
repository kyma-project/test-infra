import {
  to = google_artifact_registry_repository.dev_modules_internal
  id = "projects/kyma-project/locations/europe/repositories/dev-modules-internal"
}

# TODO (dekiel): remove after migration to modulectl is done
resource "google_artifact_registry_repository" "dev_modules_internal" {
  provider               = google.kyma_project
  location               = var.dev_modules_internal_repository.location
  repository_id          = var.dev_modules_internal_repository.name
  description            = var.dev_modules_internal_repository.description
  format                 = var.dev_modules_internal_repository.format
  cleanup_policy_dry_run = var.dev_modules_internal_repository.cleanup_policy_dry_run
  labels                 = var.dev_modules_internal_repository.labels

  docker_config {
    immutable_tags = var.dev_modules_internal_repository.immutable_tags
  }
}

import {
  id = "projects/kyma-project/serviceAccounts/kyma-submission-pipeline@kyma-project.iam.gserviceaccount.com"
  to = google_service_account.kyma_project_kyma_submission_pipeline
}

# TODO (dekiel): The submission pipeline should have only one identity in our projects.
#   The service account in kyma-project should be removed.
resource "google_service_account" "kyma_project_kyma_submission_pipeline" {
  provider     = google.kyma_project
  account_id   = "kyma-submission-pipeline"
  display_name = "kyma-submission-pipeline"
  description  = "The submission-pipeline ADO pipeline."
}

resource "google_service_account" "kyma-submission-pipeline" {
  account_id   = "kyma-submission-pipeline"
  display_name = "kyma-submission-pipeline"
  #   description = "Service account for retrieving secrets on the conduit-cli build pipeline."
  description = "The submission-pipeline ADO pipeline."
}

resource "google_artifact_registry_repository_iam_member" "dev_modules_internal_repo_admin" {
  provider   = google.kyma_project
  repository = google_artifact_registry_repository.dev_modules_internal.id
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${google_service_account.kyma_project_kyma_submission_pipeline.email}"
}

module "dev_kyma_modules" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }
  repository_prevent_destroy = var.dev_kyma_modules_repository.repository_prevent_destroy
  repository_name            = var.dev_kyma_modules_repository.name
  description                = var.dev_kyma_modules_repository.description
  repoAdmin_serviceaccounts  = [google_service_account.kyma_project_kyma_submission_pipeline.email]
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
  repoAdmin_serviceaccounts  = [google_service_account.kyma_project_kyma_submission_pipeline.email]
}
