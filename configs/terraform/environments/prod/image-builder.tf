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

resource "google_service_account" "image-builder-gh-workflow" {
  description  = "Service account used by image-builder reusable workflow to access GCP secret manager. Reusable workflow is defined in the test-infra repository."
  project      = var.image_builder_gh_workflow_service_account.project_id
  account_id   = var.image_builder_gh_workflow_service_account.id
  display_name = var.image_builder_gh_workflow_service_account.id
}

# Grant the image-builder service account the workload identity user role.
# This role is required to impersonate the image-builder-workflow GCP service account by GitHub image-builder reusable workflow using workload identity federation.
resource "google_service_account_iam_binding" "image_builder_gh_workflow_workload_identity" {
  service_account_id = google_service_account.image-builder-gh-workflow.name
  role               = "roles/iam.workloadIdentityUser"
  members = [
    "principal://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${data.github_organization.kyma-project.id}:workflow:${var.image_builder_reusable_workflow_name}"
  ]
}

# Grant read access to the GCP secret manager secret with ado pat to the image-builder service account.
# This secret is used by the image-builder reusable workflow to authenticate with Azure DevOps API and trigger ADO pipeline.
resource "google_secret_manager_secret_iam_member" "image_builder_ado_pat" {
  project   = var.gcp_project_id
  secret_id = var.image_builder_ado_pat_gcp_secret_manager_secret_name
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.image-builder-gh-workflow.email}"
}

resource "google_secret_manager_secret_iam_member" "image_builder_ado_pat" {
  project   = var.gcp_project_id
  secret_id = var.image_builder_ado_pat_gcp_secret_manager_secret_name
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.image_builder_reusable_workflow_ref}"
}

# GitHub resources

# Define GitHub Actions secrets for the image-builder reusable workflow.
# These secret contains the values of the GCP service account email used by the image-builder reusable workflow to access GCP secret manager.
resource "github_actions_variable" "image_builder_gh_workflow_gcp_service_account_email" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "IMAGE_BUILDER_GH_WORKFLOW_GCP_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.image-builder-gh-workflow.email
}

# Define GitHub Actions secrets for the image-builder reusable workflow.
# These secret contains the values of the GCP secret manager secret name with ado pat
# It's used by the image-builder reusable workflow to authenticate with Azure DevOps API and trigger ADO pipeline.
resource "github_actions_variable" "image_builder_ado_pat_gcp_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "IMAGE_BUILDER_ADO_PAT_GCP_SECRET_NAME"
  value         = var.image_builder_ado_pat_gcp_secret_manager_secret_name
}