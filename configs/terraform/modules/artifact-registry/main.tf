data "google_client_config" "this" {}

resource "google_artifact_registry_repository" "artifact_registry" {
  location      = var.multi_region == true ? var.primary_area : data.google_client_config.this.region
  repository_id = lower(var.registry_name)
  description   = "${lower(var.registry_name)} registry"
  format        = "DOCKER"

  labels = {
    name  = "${lower(var.registry_name)}"
    owner = var.owner
    type  = var.type
  }
  docker_config {
    immutable_tags = var.immutable_tags
  }
}

resource "google_artifact_registry_repository_iam_member" "writer_service_account" {
  for_each   = toset(var.writer_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = var.multi_region == true ? var.primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "reader_service_accounts" {
  for_each   = toset(var.reader_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = var.multi_region == true ? var.primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "public_access" {
  count      = var.public == true ? 1 : 0
  project    = data.google_client_config.this.project
  location   = var.multi_region == true ? var.primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "allUsers"
}# (2025-03-04)