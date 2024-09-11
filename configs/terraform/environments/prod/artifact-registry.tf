module "artifact_registry" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }


  for_each               = var.kyma_project_artifact_registry_collection
  registry_name          = each.value.name
  type                   = each.value.type
  immutable_tags         = each.value.immutable
  multi_region           = each.value.multi_region
  owner                  = each.value.owner
  writer_serviceaccounts = each.value.writer_serviceaccounts
  reader_serviceaccounts = each.value.reader_serviceaccounts
  public                 = each.value.public
}

import {
  to = google_artifact_registry_repository.prod_docker_repository
  id = "projects/${var.kyma_project_gcp_project_id}/locations/${var.prod_docker_repository.location}/repositories/${var.prod_docker_repository.name}"
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
}