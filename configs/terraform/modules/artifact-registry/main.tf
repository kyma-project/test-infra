data "google_client_config" "this" {}

# Get correct location based on multi_region flag.
locals {
  remote_repository_config = one([var.remote_repository_config])
  # This is workaround, as OpenTofu does not support yet the conditional expressions in the resource block
  # https://github.com/opentofu/opentofu/issues/1329
  repository = var.repository_prevent_destroy ? google_artifact_registry_repository.protected_repository[0] : google_artifact_registry_repository.unprotected_repository[0]
  location = var.multi_region ? (
    var.primary_area != "" ? var.primary_area : error("multi_region is true, but primary_area is not set.")
    ) : (
    var.location != "" ? var.location : error("multi_region is false, but location is not set.")
  )
}

# Resource with prevent_destroy lifecycle
resource "google_artifact_registry_repository" "protected_repository" {
  count                  = var.repository_prevent_destroy ? 1 : 0
  location               = local.location
  repository_id          = lower(var.repository_name)
  description            = var.description
  format                 = var.format
  mode                   = var.mode
  cleanup_policy_dry_run = var.cleanup_policy_dry_run

  labels = {
    name  = lower(var.repository_name)
    owner = var.owner
    type  = var.type
  }

  lifecycle {
    prevent_destroy = true
  }

  dynamic "docker_config" {
    for_each = var.mode != "REMOTE_REPOSITORY" ? [1] : []
    content {
      immutable_tags = var.immutable_tags
    }
  }

  dynamic "remote_repository_config" {
    for_each = local.remote_repository_config != null ? [local.remote_repository_config] : []
    iterator = remote_config
    content {
      description = remote_config.value.description

      docker_repository {
        public_repository = remote_config.value.docker_public_repository
      }

      dynamic "upstream_credentials" {
        for_each = (try(remote_config.value.upstream_username, null) != null &&
        try(remote_config.value.upstream_password_secret, null) != null) ? [1] : []
        content {
          username_password_credentials {
            username = remote_config.value.upstream_username
            password_secret_version = remote_config.value.upstream_password_secret
          }
        }
      }
    }
  }

  dynamic "cleanup_policies" {
    for_each = coalesce(var.cleanup_policies, [])
    iterator = policy

    content {
      id     = policy.value.id
      action = policy.value.action

      condition {
        tag_state    = try(policy.value.condition.tag_state, null)
        tag_prefixes = try(policy.value.condition.tag_prefixes, null)
        older_than   = try(policy.value.condition.older_than, null)
      }
    }
  }
}

# Resource without prevent_destroy lifecycle
resource "google_artifact_registry_repository" "unprotected_repository" {
  count                  = var.repository_prevent_destroy ? 0 : 1
  location               = local.location
  repository_id          = lower(var.repository_name)
  description            = var.description
  format                 = var.format
  mode                   = var.mode
  cleanup_policy_dry_run = var.cleanup_policy_dry_run

  labels = {
    name  = lower(var.repository_name)
    owner = var.owner
    type  = var.type
  }

  dynamic "docker_config" {
    for_each = var.mode != "REMOTE_REPOSITORY" ? [1] : []
    content {
      immutable_tags = var.immutable_tags
    }
  }

  dynamic "remote_repository_config" {
    for_each = local.remote_repository_config != null ? [local.remote_repository_config] : []
    iterator = remote_config
    content {
      description = remote_config.value.description

      docker_repository {
        public_repository = remote_config.value.docker_public_repository
      }

      dynamic "upstream_credentials" {
        for_each = (try(remote_config.value.upstream_username, null) != null &&
        try(remote_config.value.upstream_password_secret, null) != null) ? [1] : []
        content {
          username_password_credentials {
            username = remote_config.value.upstream_username
            password_secret_version = remote_config.value.upstream_password_secret
          }
        }
      }
    }
  }

  dynamic "cleanup_policies" {
    for_each = coalesce(var.cleanup_policies, [])
    iterator = policy

    content {
      id     = policy.value.id
      action = policy.value.action

      condition {
        tag_state    = try(policy.value.condition.tag_state, null)
        tag_prefixes = try(policy.value.condition.tag_prefixes, null)
        older_than   = try(policy.value.condition.older_than, null)
      }
    }
  }
}

# Updated IAM resources to reference the local.repository
resource "google_artifact_registry_repository_iam_member" "service_account_repoAdmin_access" {
  for_each   = toset(var.repoAdmin_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = local.location
  repository = local.repository.name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "service_account_writer_access" {
  for_each   = toset(var.writer_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = local.location
  repository = local.repository.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "service_account_reader_access" {
  for_each   = toset(var.reader_serviceaccounts)
  project    = data.google_client_config.this.project
  location   = local.location
  repository = local.repository.name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "group_reader_access" {
  for_each   = toset(var.reader_groups)
  project    = data.google_client_config.this.project
  location   = local.location
  repository = local.repository.name
  role       = "roles/artifactregistry.reader"
  member     = "group:${each.value}"
}

resource "google_artifact_registry_repository_iam_member" "public_access" {
  count      = var.public ? 1 : 0
  project    = data.google_client_config.this.project
  location   = local.location
  repository = local.repository.name
  role       = "roles/artifactregistry.reader"
  member     = "allUsers"
}
