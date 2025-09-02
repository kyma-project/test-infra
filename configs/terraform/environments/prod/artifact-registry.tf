module "prod_docker_repository" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.prod_docker_repository.name
  description                = var.prod_docker_repository.description
  location                   = var.prod_docker_repository.location
  immutable_tags             = var.prod_docker_repository.immutable_tags
  format                     = var.prod_docker_repository.format
  cleanup_policies           = var.prod_docker_repository.cleanup_policies
  cleanup_policy_dry_run     = var.prod_docker_repository.cleanup_policy_dry_run
  repository_prevent_destroy = var.prod_docker_repository.repository_prevent_destroy
}

module "dev_docker_repository" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.dev_docker_repository.name
  description                = var.dev_docker_repository.description
  location                   = var.dev_docker_repository.location
  immutable_tags             = var.dev_docker_repository.immutable_tags
  format                     = var.dev_docker_repository.format
  cleanup_policies           = var.dev_docker_repository.cleanup_policies
  cleanup_policy_dry_run     = var.dev_docker_repository.cleanup_policy_dry_run
  type                       = var.dev_docker_repository.type
  repository_prevent_destroy = var.dev_docker_repository.repository_prevent_destroy
}
