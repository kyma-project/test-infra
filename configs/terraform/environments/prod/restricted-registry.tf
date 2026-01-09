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

data "google_project" "kyma_project" {
  provider   = google.kyma_project
  project_id = var.kyma_project_gcp_project_id
}

import {
  to = google_secret_manager_secret.chainguard_pull_token
  id = "projects/sap-kyma-prow/secrets/chainguard_auth_token"
}

resource "google_secret_manager_secret" "chainguard_pull_token" {
  project   = var.gcp_project_id
  secret_id = var.chainguard_pull_token_secret_name

  replication {
    auto {}
  }

  labels = {
    type      = "authentication-token"
    tool      = "chainguard"
    owner     = "neighbors"
    component = "restricted-registry"
  }
}

data "google_secret_manager_secret_version" "chainguard_pull_token_password" {
  provider = google.kyma_project
  project  = var.gcp_project_id
  secret   = google_secret_manager_secret.chainguard_pull_token.secret_id
}

# Grant Terraform planner service account read access to Chainguard auth secret
resource "google_secret_manager_secret_iam_member" "chainguard_token_terraform_planner_access" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.chainguard_pull_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.terraform-planner.email}"
}

# Grant Artifact Registry service account access to the Chainguard auth secret
resource "google_secret_manager_secret_iam_member" "chainguard_token_artifactregistry_access" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.chainguard_pull_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:service-${data.google_project.kyma_project.number}@gcp-sa-artifactregistry.iam.gserviceaccount.com"
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

  remote_repository_config = {
    description               = var.chainguard_cache.remote_repository_config.description
    docker_custom_repository  = var.chainguard_cache.remote_repository_config.docker_custom_repository
    upstream_username         = "ac2ef7e1fdf69e1bf40cd31e4a868af7cca02037/034e2363347226e7"
    upstream_password_secret  = data.google_secret_manager_secret_version.chainguard_pull_token_password.name
  }
}

module "restricted_prod" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.restricted_prod.name
  description                = var.restricted_prod.description
  location                   = var.restricted_prod.location
  format                     = var.restricted_prod.format
  mode                       = var.restricted_prod.mode
  type                       = var.restricted_prod.type
  cleanup_policy_dry_run     = var.restricted_prod.cleanup_policy_dry_run
  repository_prevent_destroy = var.restricted_prod.repository_prevent_destroy
  reader_groups              = [var.restricted_registry_iam_groups.prod_read]
  virtual_repository_config  = var.restricted_prod.virtual_repository_config

  depends_on = [
    module.kyma_restricted_images_prod,
    module.chainguard_cache
  ]
}

module "restricted_dev" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name            = var.restricted_dev.name
  description                = var.restricted_dev.description
  location                   = var.restricted_dev.location
  format                     = var.restricted_dev.format
  mode                       = var.restricted_dev.mode
  type                       = var.restricted_dev.type
  cleanup_policy_dry_run     = var.restricted_dev.cleanup_policy_dry_run
  repository_prevent_destroy = var.restricted_dev.repository_prevent_destroy
  reader_groups              = [var.restricted_registry_iam_groups.dev_read]
  virtual_repository_config  = var.restricted_dev.virtual_repository_config

  depends_on = [
    module.kyma_restricted_images_dev,
    module.chainguard_cache
  ]
}

resource "google_artifact_registry_repository_iam_member" "restricted_prod_writers" {
  provider   = google.kyma_project
  project    = "kyma-project"
  location   = module.restricted_prod.artifact_registry.location
  repository = module.restricted_prod.artifact_registry.name
  role       = "roles/artifactregistry.writer"
  member     = "group:${var.restricted_registry_iam_groups.prod_write}"
}

resource "google_artifact_registry_repository_iam_member" "restricted_dev_writers" {
  provider   = google.kyma_project
  project    = "kyma-project"
  location   = module.restricted_dev.artifact_registry.location
  repository = module.restricted_dev.artifact_registry.name
  role       = "roles/artifactregistry.writer"
  member     = "group:${var.restricted_registry_iam_groups.dev_write}"
}
