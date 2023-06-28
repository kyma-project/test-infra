

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

resource "google_artifact_registry_repository_iam_member" "member_service_account" {
  count      = var.artifact_registry_count
  project    = var.gcp_project_id
  location   = local.location
  repository = google_artifact_registry_repository.artifact_registry[count.index].name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:sa-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
}