data "google_client_config" "this" {}

resource "google_artifact_registry_repository" "artifact_registry" {
  location      = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository_id = "modules-${lower(var.artifact_registry_name)}"
  description   = "modules-${lower(var.artifact_registry_name)} repository"
  format        = "DOCKER"

  labels = {
    name   = "modules-${lower(var.artifact_registry_name)}"
    owner  = var.artifact_registry_owner
    module = var.artifact_registry_module
    type   = var.artifact_registry_type
  }
  docker_config {
    immutable_tags = var.immutable_artifact_registry
  }
}

resource "google_artifact_registry_repository_iam_member" "member_service_account" {
  project    = data.google_client_config.this.project
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${var.artifact_registry_serviceaccount}"
}