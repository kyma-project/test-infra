# GCP resources

locals {
  image_builder_azure_sp_gcp_secrets = toset(values(var.image_builder_azure_sp_gcp_secret_names))
}

resource "google_secret_manager_secret_iam_member" "image_builder_public_github_reusable_workflow_principal_azure_sp_secret_reader" {
  for_each  = local.image_builder_azure_sp_gcp_secrets
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.image_builder_azure_sp[each.key].secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.image_builder_reusable_workflow_ref}"
}

resource "google_secret_manager_secret_iam_member" "image_builder_internal_github_reusable_workflow_principal_azure_sp_secret_reader" {
  for_each  = local.image_builder_azure_sp_gcp_secrets
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.image_builder_azure_sp[each.key].secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${local.internal_github_wif_pool_name}/attribute.reusable_workflow_ref/${var.image_builder_internal_github_reusable_workflow_ref}"
}

# GitHub resources

# Define GitHub Actions secrets for the image-builder reusable workflow.
# These secrets contain the values of the GCP secret manager secret names with Azure SP credentials.
# It's used by the image-builder reusable workflow to authenticate with Azure DevOps API and trigger ADO pipeline.
resource "github_actions_organization_variable" "image_builder_azure_sp_tenant_id_gcp_secret_name" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_AZURE_SP_TENANT_ID_GCP_SECRET_NAME"
  value         = var.image_builder_azure_sp_gcp_secret_names.tenant_id
}

resource "github_actions_organization_variable" "image_builder_azure_sp_tenant_id_gcp_secret_name_internal_github" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_AZURE_SP_TENANT_ID_GCP_SECRET_NAME"
  value         = var.image_builder_azure_sp_gcp_secret_names.tenant_id
}

resource "github_actions_organization_variable" "image_builder_azure_sp_client_id_gcp_secret_name" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_AZURE_SP_CLIENT_ID_GCP_SECRET_NAME"
  value         = var.image_builder_azure_sp_gcp_secret_names.client_id
}

resource "github_actions_organization_variable" "image_builder_azure_sp_client_id_gcp_secret_name_internal_github" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_AZURE_SP_CLIENT_ID_GCP_SECRET_NAME"
  value         = var.image_builder_azure_sp_gcp_secret_names.client_id
}

resource "github_actions_organization_variable" "image_builder_azure_sp_client_secret_gcp_secret_name" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_AZURE_SP_CLIENT_SECRET_GCP_SECRET_NAME"
  value         = var.image_builder_azure_sp_gcp_secret_names.client_secret
}

resource "github_actions_organization_variable" "image_builder_azure_sp_client_secret_gcp_secret_name_internal_github" {
  provider      = github.internal_github
  visibility    = "all"
  variable_name = "IMAGE_BUILDER_AZURE_SP_CLIENT_SECRET_GCP_SECRET_NAME"
  value         = var.image_builder_azure_sp_gcp_secret_names.client_secret
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

# Secrets in sap-kyma-prow project to store Azure SP credentials (to be added manually)
resource "google_secret_manager_secret" "image_builder_azure_sp" {
  for_each  = local.image_builder_azure_sp_gcp_secrets
  project   = var.gcp_project_id
  secret_id = each.value

  replication {
    auto {}
  }

  labels = {
    type      = "azure-sp-credential"
    tool      = "image-builder"
    owner     = "neighbors"
    component = "oci-image-builder"
  }
}

# Secrets used by kyma-oci-image-builder SA in the ADO pipeline.
import {
  to = google_secret_manager_secret.oci_image_builder_azure_pipeline_sa
  id = "projects/${var.gcp_project_id}/secrets/azure-pipeline-image-builder-sa"
}

resource "google_secret_manager_secret" "oci_image_builder_azure_pipeline_sa" {
  project   = var.gcp_project_id
  secret_id = "azure-pipeline-image-builder-sa"

  replication {
    auto {}
  }

  labels = {
    type      = "service-account-key"
    tool      = "image-builder"
    owner     = "neighbors"
    component = "oci-image-builder"
  }
}

import {
  to = google_secret_manager_secret.oci_image_builder_sa_key_restricted_markets
  id = "projects/${var.gcp_project_id}/secrets/image-builder-sa-key-restricted-markets"
}

resource "google_secret_manager_secret" "oci_image_builder_sa_key_restricted_markets" {
  project   = var.gcp_project_id
  secret_id = "image-builder-sa-key-restricted-markets"

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

import {
  to = google_secret_manager_secret.oci_image_builder_signify_prod
  id = "projects/${var.gcp_project_id}/secrets/kyma-signify-prod"
}

resource "google_secret_manager_secret" "oci_image_builder_signify_prod" {
  project   = var.gcp_project_id
  secret_id = "kyma-signify-prod"

  replication {
    auto {}
  }

  labels = {
    type      = "signify"
    tool      = "image-builder"
    owner     = "neighbors"
    component = "oci-image-builder"
    usage     = "image-builder"
  }

  lifecycle {
    ignore_changes = [rotation, topics]
  }
}

import {
  to = google_secret_manager_secret.oci_image_builder_sap_github_prow_sa_token
  id = "projects/${var.gcp_project_id}/secrets/kyma-sap-github-prow-sa-token"
}

resource "google_secret_manager_secret" "oci_image_builder_sap_github_prow_sa_token" {
  project   = var.gcp_project_id
  secret_id = "kyma-sap-github-prow-sa-token"

  replication {
    auto {}
  }

  labels = {
    type            = "github-token"
    tool            = "image-builder"
    github-instance = "internal"
    owner           = "neighbors"
    component       = "oci-image-builder"
  }
}

# Per-secret IAM bindings for kyma-oci-image-builder SA.
# Replaces the overly broad project-level roles/secretmanager.secretAccessor binding.
import {
  to = google_secret_manager_secret_iam_member.oci_image_builder_azure_pipeline_sa_secret_accessor
  id = "projects/${var.gcp_project_id}/secrets/azure-pipeline-image-builder-sa/members/serviceAccount:kyma-oci-image-builder@${var.gcp_project_id}.iam.gserviceaccount.com/roles/secretmanager.secretAccessor"
}

resource "google_secret_manager_secret_iam_member" "oci_image_builder_azure_pipeline_sa_secret_accessor" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.oci_image_builder_azure_pipeline_sa.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kyma-oci-image-builder.email}"
}

import {
  to = google_secret_manager_secret_iam_member.oci_image_builder_sa_key_restricted_markets_secret_accessor
  id = "projects/${var.gcp_project_id}/secrets/image-builder-sa-key-restricted-markets/members/serviceAccount:kyma-oci-image-builder@${var.gcp_project_id}.iam.gserviceaccount.com/roles/secretmanager.secretAccessor"
}

resource "google_secret_manager_secret_iam_member" "oci_image_builder_sa_key_restricted_markets_secret_accessor" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.oci_image_builder_sa_key_restricted_markets.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kyma-oci-image-builder.email}"
}

import {
  to = google_secret_manager_secret_iam_member.oci_image_builder_signify_prod_secret_accessor
  id = "projects/${var.gcp_project_id}/secrets/kyma-signify-prod/members/serviceAccount:kyma-oci-image-builder@${var.gcp_project_id}.iam.gserviceaccount.com/roles/secretmanager.secretAccessor"
}

resource "google_secret_manager_secret_iam_member" "oci_image_builder_signify_prod_secret_accessor" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.oci_image_builder_signify_prod.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kyma-oci-image-builder.email}"
}

import {
  to = google_secret_manager_secret_iam_member.oci_image_builder_sap_github_prow_sa_token_secret_accessor
  id = "projects/${var.gcp_project_id}/secrets/kyma-sap-github-prow-sa-token/members/serviceAccount:kyma-oci-image-builder@${var.gcp_project_id}.iam.gserviceaccount.com/roles/secretmanager.secretAccessor"
}

resource "google_secret_manager_secret_iam_member" "oci_image_builder_sap_github_prow_sa_token_secret_accessor" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.oci_image_builder_sap_github_prow_sa_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kyma-oci-image-builder.email}"
}
