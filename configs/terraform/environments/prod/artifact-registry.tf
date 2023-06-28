

resource "google_artifact_registry_repository" "artifact_registry" {
  count         = var.artifact_registry_count
  location      = local.location
  repository_id = "${lower(var.artifact_registry_prefix)}-${lower(var.artifact_registry_names[count.index])}"
  description   = "${lower(var.artifact_registry_prefix)}-${lower(var.artifact_registry_names[count.index])} repository"
  format        = "DOCKER"

  labels = merge(local.artifact_registry_tags, {
    name = "${lower(var.artifact_registry_prefix)}-${lower(var.artifact_registry_names[count.index])}"
  })

  docker_config {
    immutable_tags = var.immutable_artifact_registry
  }
}

resource "google_artifact_registry_repository_iam_member" "member_service_account_1" {
  count      = var.artifact_registry_count
  project    = var.gcp_project
  location   = local.location
  repository = google_artifact_registry_repository.artifact_registry[count.index].name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:sa-dev-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
}

resource "google_artifact_registry_repository_iam_member" "member_service_account_2" {
  count      = var.artifact_registry_count
  project    = var.gcp_project
  location   = local.location
  repository = google_artifact_registry_repository.artifact_registry[count.index].name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:sa-gcr-push-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
}