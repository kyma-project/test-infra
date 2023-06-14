# Secure access to signify dev and prod secrets over k8s API.
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
    namespace = var.external_secrets_sa_trusted_cluster.namespace
    name      = var.external_secrets_sa_trusted_cluster.name
    api_group = "rbac.authorization.k8s.io"
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
    namespace = var.external_secrets_sa_trusted_cluster.namespace
    name      = var.external_secrets_sa_trusted_cluster.name
    api_group = "rbac.authorization.k8s.io"
  }
}
