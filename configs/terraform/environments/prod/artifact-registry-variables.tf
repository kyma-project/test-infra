###################################
# Artifact Registry related values
###################################
variable "kyma_project_artifact_registry_collection" {
  description = "Artifact Registry related data set"
  type = map(object({
    name                   = string
    owner                  = string
    type                   = string
    writer_serviceaccounts = optional(list(string), [])
    reader_serviceaccounts = list(string)
    primary_area           = optional(string, "europe")
    multi_region           = optional(bool, true)
    public                 = optional(bool, false)
    immutable              = optional(bool, false)
  }))
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
  format                 = var.prod_docker_repository.format
  cleanup_policy_dry_run = var.prod_docker_repository.cleanup_policy_dry_run
  docker_config {
    immutable_tags = var.prod_docker_repository.immutable_tags
  }
}