# Secure access to pjtester secret over k8s API.
resource "kubernetes_cluster_role" "access_pjtester_secrets_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-pjtester-secrets"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [var.pjtester_kubeconfig_secret_name, var.pjtester_github_token_secret_name]
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
    name      = kubernetes_cluster_role.access_pjtester_secrets_trusted_workloads.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external_secrets_sa_trusted_cluster.namespace
    name      = var.external_secrets_sa_trusted_cluster.name
    api_group = "rbac.authorization.k8s.io"
  }
}
