

resource "google_artifact_registry_repository" "artifact_registry" {
  count         = var.artifact_registry_count
  location      = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : var.gcp_region
  repository_id = "${lower(var.artifact_registry_prefix)}-${lower(var.artifact_registry_names[count.index])}"
  description   = "${lower(var.artifact_registry_prefix)}-${lower(var.artifact_registry_names[count.index])} repository"
  format        = "DOCKER"

  labels = {
    name   = "${lower(var.artifact_registry_prefix)}-${lower(var.artifact_registry_names[count.index])}"
    owner  = var.artifact_registry_owner
    module = var.artifact_registry_module
    type   = var.artifact_registry_type
  }
  docker_config {
    immutable_tags = var.immutable_artifact_registry
  }
}

resource "google_artifact_registry_repository_iam_member" "member_service_account" {
  count      = var.artifact_registry_count
  project    = var.gcp_project_id
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : var.gcp_region
  repository = google_artifact_registry_repository.artifact_registry[count.index].name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:sa-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
}