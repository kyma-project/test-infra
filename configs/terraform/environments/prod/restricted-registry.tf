module "kyma_restricted_images_prod" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.kyma_restricted_images_prod.name
  description                = var.kyma_restricted_images_prod.description
  location                   = var.kyma_restricted_images_prod.location
  immutable_tags             = var.kyma_restricted_images_prod.immutable_tags
  format                     = var.kyma_restricted_images_prod.format
  mode                       = var.kyma_restricted_images_prod.mode
  type                       = var.kyma_restricted_images_prod.type
  cleanup_policies           = var.kyma_restricted_images_prod.cleanup_policies
  cleanup_policy_dry_run     = var.kyma_restricted_images_prod.cleanup_policy_dry_run
  repository_prevent_destroy = var.kyma_restricted_images_prod.repository_prevent_destroy
}

module "kyma_restricted_images_dev" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.kyma_restricted_images_dev.name
  description                = var.kyma_restricted_images_dev.description
  location                   = var.kyma_restricted_images_dev.location
  immutable_tags             = var.kyma_restricted_images_dev.immutable_tags
  format                     = var.kyma_restricted_images_dev.format
  mode                       = var.kyma_restricted_images_dev.mode
  type                       = var.kyma_restricted_images_dev.type
  cleanup_policies           = var.kyma_restricted_images_dev.cleanup_policies
  cleanup_policy_dry_run     = var.kyma_restricted_images_dev.cleanup_policy_dry_run
  repository_prevent_destroy = var.kyma_restricted_images_dev.repository_prevent_destroy
}

module "chainguard_cache" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.chainguard_cache.name
  description                = var.chainguard_cache.description
  location                   = var.chainguard_cache.location
  format                     = var.chainguard_cache.format
  mode                       = var.chainguard_cache.mode
  type                       = var.chainguard_cache.type
  cleanup_policy_dry_run     = var.chainguard_cache.cleanup_policy_dry_run
  repository_prevent_destroy = var.chainguard_cache.repository_prevent_destroy
  remote_repository_config   = var.chainguard_cache.remote_repository_config
}
