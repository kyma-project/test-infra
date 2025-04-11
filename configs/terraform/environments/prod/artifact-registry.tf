module "artifact_registry" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }


  for_each               = var.kyma_project_artifact_registry_collection
  repository_name        = each.value.name
  description            = each.value.description
  type                   = each.value.type
  immutable_tags         = each.value.immutable
  multi_region           = each.value.multi_region
  owner                  = each.value.owner
  repoAdmin_serviceaccounts = each.value.repoAdmin_serviceaccounts
  reader_serviceaccounts = each.value.reader_serviceaccounts
  public                 = each.value.public
  cleanup_policy_dry_run = each.value.cleanup_policy_dry_run
  cleanup_policies       = each.value.cleanup_policies
}

resource "google_artifact_registry_repository" "prod_docker_repository" {
  provider               = google.kyma_project
  labels                 = var.prod_docker_repository.labels
  location               = var.prod_docker_repository.location
  repository_id          = var.prod_docker_repository.name
  description            = var.prod_docker_repository.description
  format                 = var.prod_docker_repository.format
  cleanup_policy_dry_run = var.prod_docker_repository.cleanup_policy_dry_run
  docker_config {
    immutable_tags = var.prod_docker_repository.immutable_tags
  }

  cleanup_policies {
    id     = "delete-untagged"
    action = "DELETE"
    condition {
      tag_state = "UNTAGGED"
    }
  }
}

resource "google_artifact_registry_repository" "docker_dev" {
  provider               = google.kyma_project
  location               = var.docker_dev_repository.location
  repository_id          = var.docker_dev_repository.name
  description            = var.docker_dev_repository.description
  format                 = var.docker_dev_repository.format
  cleanup_policy_dry_run = var.docker_dev_repository.cleanup_policy_dry_run

  docker_config {
    immutable_tags = var.docker_dev_repository.immutable_tags
  }

  cleanup_policies {
    id     = "delete-untagged"
    action = "DELETE"
    condition {
      tag_state = "UNTAGGED"
    }
  }

  cleanup_policies {
    id     = "delete-old-pr-images"
    action = "DELETE"
    condition {
      tag_state = "TAGGED"
      # Equivalent to PR-*
      tag_prefixes = ["PR-"]
      older_than   = var.docker_dev_repository.pr_images_max_age
    }
  }

  labels = var.docker_dev_repository.labels
}