data "google_client_config" "this" {}

# Get correct location based on multi_region flag.
locals {
  location = var.multi_region ? (
    var.primary_area != "" ? var.primary_area : error("multi_region is true, but primary_area is not set.")
  ) : (
    var.location != "" ? var.location : error("multi_region is false, but location is not set.")
  )
}

resource "google_artifact_registry_repository" "artifact_registry" {
  location               = local.location
  repository_id = lower(var.repository_name)
  description            = var.description
  format                 = var.format
  cleanup_policy_dry_run = var.cleanup_policy_dry_run

  labels = {
    name = lower(var.repository_name)
    owner = var.owner
    type  = var.type
  }

  docker_config {
    immutable_tags = var.immutable_tags
  }

  dynamic "cleanup_policies" {
    for_each = coalesce(var.cleanup_policies, [])
    iterator = policy

    content {
      id     = policy.value.id
      action = policy.value.action

      condition {
        tag_state = try(policy.value.condition.tag_state, null)
        tag_prefixes = try(policy.value.condition.tag_prefixes, null)
        older_than = try(policy.value.condition.older_than, null)
      }
    }

  }
}

resource "google_artifact_registry_repository_iam_member" "service_account_repoAdmin_access" {
  for_each = toset(var.repoAdmin_serviceaccounts)
  project    = data.google_client_config.this.project
  location = local.location
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "service_account_writer_access" {
  for_each = toset(var.writer_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = local.location
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "service_account_reader_access" {
  for_each   = toset(var.reader_serviceaccounts)
  project    = data.google_client_config.this.project
  location = local.location
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "public_access" {
  count    = var.public ? 1 : 0
  project    = data.google_client_config.this.project
  location = local.location
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.reader"
  member     = "allUsers"
}