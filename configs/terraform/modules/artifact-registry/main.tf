

resource "google_artifact_registry_repository" "artifact_registry" {
  for_each      = toset(var.artifact_registry_names)
  location      = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : var.gcp_region
  repository_id = "modules-${lower(each.value)}"
  description   = "modules-${lower(each.value)} repository"
  format        = "DOCKER"

  labels = {
    name   = "modules-${lower(each.value)}"
    owner  = var.artifact_registry_owner
    module = var.artifact_registry_module
    type   = var.artifact_registry_type
  }
  docker_config {
    immutable_tags = var.immutable_artifact_registry
  }
}

resource "google_artifact_registry_repository_iam_member" "member_service_account" {
  for_each   = google_artifact_registry_repository.artifact_registry
  project    = var.gcp_project_id
  location   = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : var.gcp_region
  repository = google_artifact_registry_repository.artifact_registry[each.key].name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:sa-kyma-project@sap-kyma-prow.iam.gserviceaccount.com"
}