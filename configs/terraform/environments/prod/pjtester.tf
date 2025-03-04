# Secure access to pjtester secret over k8s API.
# Only external-secrets controller need access to this secret over k8s API.
# Prowjobs access this secret as env vars or mounted files. This is controlled by OPA Gatekeeper.

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
    namespace = var.external_secrets_k8s_sa_trusted_cluster.namespace
    name      = var.external_secrets_k8s_sa_trusted_cluster.name
  }
}
# (2025-03-04)