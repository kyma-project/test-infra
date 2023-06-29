# Secure access to signify dev and prod secrets over k8s API.
# Only external-secrets controller need access to these secrets over k8s API.
# Prowjobs access these secrets as env vars or mounted files. This is controlled by OPA Gatekeeper.

resource "kubernetes_cluster_role" "access_kyma_autobump_bot_github_token_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-kyma-autobump-bot-github-token"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [var.kyma_autobump_bot_github_token_secret_name]
    verbs          = ["update", "get", "list", "watch", "patch", "create", "delete"]
  }
}

resource "kubernetes_cluster_role" "access_kyma_autobump_bot_github_token_untrusted_workloads" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    name = "access-kyma-autobump-bot-github-token"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [var.kyma_autobump_bot_github_token_secret_name]
    verbs          = ["update", "get", "list", "watch", "patch", "create", "delete"]
  }
}

resource "kubernetes_cluster_role_binding" "access_kyma_autobump_bot_github_token_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-access-kyma-autobump-bot-github-token"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.access_kyma_autobump_bot_github_token_trusted_workloads.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external_secrets_k8s_sa_trusted_cluster.namespace
    name      = var.external_secrets_k8s_sa_trusted_cluster.name
  }
}
resource "kubernetes_cluster_role_binding" "access_kyma_autobump_bot_github_token_untrusted_workloads" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    name = "access-access-kyma-autobump-bot-github-token"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.access_kyma_autobump_bot_github_token_untrusted_workloads.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external_secrets_k8s_sa_trusted_cluster.namespace
    name      = var.external_secrets_k8s_sa_trusted_cluster.name
  }
}
