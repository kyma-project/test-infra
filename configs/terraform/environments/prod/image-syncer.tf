resource "google_service_account" "image_syncer_reader" {
  account_id  = var.image_syncer_reader_service_account_name
  description = "Service account for image-syncer github reusable workflow called on pull request event. This service account is used to pull images from the prod Docker repository. It must not have the ability to push images to the prod Docker repository."
}

resource "google_service_account_iam_member" "image_syncer_reader_workflow_sa_user" {
  service_account_id = google_service_account.image_syncer_reader.name
  role = "roles/iam.workloadIdentityUser"
  member = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:pull_request_target:repository_owner_id:${data.github_organization.kyma-project.id}:reusable_workflow_ref:${var.image_syncer_reusable_workflow_ref}"
}

resource "google_artifact_registry_repository_iam_member" "image_syncer_prod_repo_writer" {
  provider = google.kyma_project
  location   = google_artifact_registry_repository.prod_docker_repository.location
  repository = google_artifact_registry_repository.prod_docker_repository.name
  role       = "roles/artifactregistry.createOnPushWriter"
  member     = "serviceAccount:${google_service_account.image_syncer_writer.email}"
}

resource "google_service_account" "image_syncer_writer" {
  account_id  = var.image_syncer_writer_service_account_name
  description = "Service account for image-syncer github reusable workflow called on push event. This service account is used to push images to the prod Docker repository."
}

resource "google_service_account_iam_member" "image_syncer_writer_workflow_sa_user" {
  service_account_id = google_service_account.image_syncer_writer.name
  role = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:push:repository_owner_id:${data.github_organization.kyma-project.id}:reusable_workflow_ref:${var.image_syncer_reusable_workflow_ref}"
}

resource "google_artifact_registry_repository_iam_member" "image_syncer_prod_repo_reader" {
  provider = google.kyma_project
  location   = google_artifact_registry_repository.prod_docker_repository.location
  repository = google_artifact_registry_repository.prod_docker_repository.name
  role     = "roles/artifactregistry.reader"
  member   = "serviceAccount:${google_service_account.image_syncer_reader.email}"
}

resource "github_actions_organization_variable" "image_syncer_reader_service_account_email" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "IMAGE_SYNCER_READER_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.image_syncer_reader.email
}

resource "github_actions_organization_variable" "image_syncer_writer_service_account_email" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "IMAGE_SYNCER_WRITER_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.image_syncer_writer.email
}