# Create the secret with the terraform executor service account token.
resource "kubernetes_secret" "terraform_executor" {
  metadata {
    name      = var.terraform_executor_k8s_service_account.name
    namespace = var.terraform_executor_k8s_service_account.namespace
    annotations = {
      "kubernetes.io/service-account.name" = kubernetes_service_account.terraform_executor.metadata[0].name
    }
  }
  type = "kubernetes.io/service-account-token"
}


# Create the terraform executor k8s service account with annotations for workload identity.
resource "kubernetes_service_account" "terraform_executor" {
  metadata {
    namespace = var.terraform_executor_k8s_service_account.namespace
    name      = var.terraform_executor_k8s_service_account.name
    annotations = {
      "iam.gke.io/gcp-service-account" = format("%s@%s.iam.gserviceaccount.com", var
      .terraform_executor_gcp_service_account.id, var.terraform_executor_gcp_service_account.project_id)
    }
  }
  automount_service_account_token = true
}

# Secure access to terraform-executor secret over k8s API.
resource "kubernetes_cluster_role" "access_terraform_executor_secret" {

  metadata {
    name = "access-terraform-executor-secret"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [kubernetes_secret.terraform_executor.metadata[0].name]
    verbs          = ["update", "get", "list", "watch", "patch", "create", "delete"]
  }
}

resource "kubernetes_cluster_role_binding" "access_pjtester_secrets_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-pjtester-secrets"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.access_terraform_executor_secret.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_secret.terraform_executor.metadata[0].name
    namespace = kubernetes_secret.terraform_executor.metadata[0].namespace
  }
  subject {
    kind      = "ServiceAccount"
    name      = var.external_secrets_sa.name
    namespace = var.external_secrets_sa.namespace
    api_group = "rbac.authorization.k8s.io"
  }
}
