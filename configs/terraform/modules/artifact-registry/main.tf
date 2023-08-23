data "google_client_config" "this" {}

resource "google_artifact_registry_repository" "artifact_registry" {
  location      = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository_id = lower(var.artifact_registry_name)
  description   = "${lower(var.artifact_registry_name)} repository"
  format        = "DOCKER"

  labels = {
    name  = "${lower(var.artifact_registry_name)}"
    owner = var.artifact_registry_owner
    type  = var.artifact_registry_type
  }
  docker_config {
    immutable_tags = var.artifact_registry_immutable_tags
  }
}

resource "google_artifact_registry_repository_iam_member" "writer_service_account" {
  for_each   = toset(var.artifact_registry_writer_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "reader_service_accounts" {
  for_each   = toset(var.artifact_registry_reader_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "public_access" {
  count      = var.artifact_registry_public == true ? 1 : 0
  project    = data.google_client_config.this.project
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "allUsers"
}