data "google_client_config" "this" {}

resource "google_artifact_registry_repository" "artifact_registry" {
  location      = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository_id = lower(var.artifact_registry_name)
  description   = "${lower(var.artifact_registry_name)} repository"
  format        = "DOCKER"

  labels = {
    name   = "${lower(var.artifact_registry_name)}"
    owner  = var.artifact_registry_owner
    module = var.artifact_registry_module
    type   = var.artifact_registry_type
  }
  docker_config {
    immutable_tags = var.immutable_artifact_registry
  }
}

resource "google_artifact_registry_repository_iam_member" "member_service_account" {
  count      = var.artifact_registry_writer_serviceaccount == "" ? 0 : 1
  project    = data.google_client_config.this.project
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${var.artifact_registry_writer_serviceaccount}"
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