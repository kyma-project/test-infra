# GCP resources

resource "google_secret_manager_secret_iam_member" "image_builder_reusable_workflow_principal_ado_pat_reader" {
  project   = var.gcp_project_id
  secret_id = var.image_builder_ado_pat_gcp_secret_manager_secret_name
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.image_builder_reusable_workflow_ref}"
}

# GitHub resources

# Define GitHub Actions secrets for the image-builder reusable workflow.
# These secret contains the values of the GCP secret manager secret name with ado pat
# It's used by the image-builder reusable workflow to authenticate with Azure DevOps API and trigger ADO pipeline.
resource "github_actions_organization_variable" "image_builder_ado_pat_gcp_secret_name" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_ADO_PAT_GCP_SECRET_NAME"
  value         = var.image_builder_ado_pat_gcp_secret_manager_secret_name
}

# This resource will be destroyed and created in case of any changes. This is not a crucial for this resource.
module "dockerhub_mirror" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_prevent_destroy = var.dockerhub_mirror.repository_prevent_destroy
  repository_name            = var.dockerhub_mirror.name
  description                = var.dockerhub_mirror.description
  repoAdmin_serviceaccounts  = [google_service_account.kyma_project_image_builder.email]
  location                   = var.dockerhub_mirror.location
  format                     = var.dockerhub_mirror.format
  mode                       = var.dockerhub_mirror.mode
  cleanup_policy_dry_run     = var.dockerhub_mirror.cleanup_policy_dry_run

  remote_repository_config = {
    description = var.dockerhub_mirror.remote_repository_config.description
    docker_public_repository = var.dockerhub_mirror.remote_repository_config.docker_public_repository
    upstream_username = var.dockerhub_credentials != null ? var.dockerhub_credentials.username : null
    upstream_password_secret = data.google_secret_manager_secret_version.dockerhub_oat_secret[0].name
  }
}

# This resource will be destroyed and created in case of any changes. This is not a crucial for this resource.
module "docker_cache" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_prevent_destroy = var.docker_cache_repository.repository_prevent_destroy
  repository_name            = var.docker_cache_repository.name
  description                = var.docker_cache_repository.description
  repoAdmin_serviceaccounts  = [
    google_service_account.kyma_project_image_builder.email,
    google_service_account.kyma_project_image_builder_restricted_markets.email
  ]
  location                   = var.docker_cache_repository.location
  format                     = var.docker_cache_repository.format
  immutable_tags             = var.docker_cache_repository.immutable_tags
  mode                       = var.docker_cache_repository.mode
  cleanup_policy_dry_run     = var.docker_cache_repository.cleanup_policy_dry_run
  cleanup_policies           = var.docker_cache_repository.cleanup_policies
}

resource "google_service_account" "kyma_project_image_builder" {
  provider    = google.kyma_project
  account_id  = var.image_builder_kyma-project_identity.id
  description = var.image_builder_kyma-project_identity.description
}

# Restricted markets image-builder service account
resource "google_service_account" "kyma_project_image_builder_restricted_markets" {
  provider    = google.kyma_project
  account_id  = var.image_builder_kyma-project_identity_restricted_markets.id
  description = var.image_builder_kyma-project_identity_restricted_markets.description
}

# Secret in sap-kyma-prow project to store the service account key (to be added manually)
resource "google_secret_manager_secret" "image_builder_sa_key_restricted_markets" {
  project   = var.gcp_project_id
  secret_id = var.image_builder_sa_key_restricted_markets_secret_name

  replication {
    auto {}
  }

  labels = {
    type        = "service-account-key"
    tool        = "image-builder"
    environment = "restricted-markets"
    owner       = "neighbors"
    component   = "oci-image-builder"
  }
}

