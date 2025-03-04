import {
  to = google_artifact_registry_repository.dev_modules_internal
  id = "projects/kyma-project/locations/europe/repositories/dev-modules-internal"
}

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

# The submission pipeline should have only one identity in our projects.
# The service account in kyma-project should be removed.
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
}# (2025-03-04)