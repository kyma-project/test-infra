# Secure access to signify dev and prod secrets over k8s API.
# Only external-secrets controller need access to these secrets over k8s API.
# Prowjobs access these secrets as env vars or mounted files. This is controlled by OPA Gatekeeper.

resource "kubernetes_cluster_role" "access_signify_secrets_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [var.signify_dev_secret_name, var.signify_prod_secret_name]
    verbs          = ["update", "get", "list", "watch", "patch", "create", "delete"]
  }
}

resource "kubernetes_cluster_role" "access_signify_secrets_untrusted_workloads" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [var.signify_dev_secret_name, var.signify_prod_secret_name]
    verbs          = ["update", "get", "list", "watch", "patch", "create", "delete"]
  }
}

resource "kubernetes_cluster_role_binding" "access_signify_prod_secret_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.access_signify_secrets_trusted_workloads.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external_secrets_k8s_sa_trusted_cluster.namespace
    name      = var.external_secrets_k8s_sa_trusted_cluster.name
  }
}
resource "kubernetes_cluster_role_binding" "access_signify_prod_secret_untrusted_workloads" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.access_signify_secrets_untrusted_workloads.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external_secrets_k8s_sa_trusted_cluster.namespace
    name      = var.external_secrets_k8s_sa_trusted_cluster.name
  }
}

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
  repoAdmin_serviceaccounts  = [google_service_account.kyma-oci-image-builder.email]
  location                   = var.dockerhub_mirror.location
  format                     = var.dockerhub_mirror.format
  immutable_tags             = var.dockerhub_mirror.immutable_tags
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
  repoAdmin_serviceaccounts  = [google_service_account.kyma-oci-image-builder.email]
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
