# TODO: Iterate over the list and create the resources for each cluster.


# Create the terraform executor k8s service account with annotations for workload identity.
resource "kubernetes_service_account" "image_builder_trusted_workloads" {
  provider = "kubernetes.trusted_workload_k8s_cluster"

  metadata {
    namespace = var.image_builder_k8s_service_account.namespace
    name      = var.image_builder_k8s_service_account.name
  }
  automount_service_account_token = true
}

resource "kubernetes_service_account" "image_builder_untrusted_workloads" {
  provider = "kubernetes.untrusted_workload_k8s_cluster"

  metadata {
    namespace = var.image_builder_k8s_service_account.namespace
    name      = var.image_builder_k8s_service_account.name
  }
  automount_service_account_token = true
}

resource "kubernetes_cluster_role" "access_signify_secrets_trusted_workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }

  rule {
    api_groups     = [""]
    resources      = ["secrets"]
    resource_names = [var.signify-dev-secret-name, var.signify-prod-secret-name]
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
    resource_names = [var.signify-dev-secret-name, var.signify-prod-secret-name]
    verbs          = ["update", "get", "list", "watch", "patch", "create", "delete"]
  }
}

resource "kubernetes_cluster_role_binding" "access-signify-prod-secret-trusted-workloads" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "access-signify-secrets"
  }
  subject {
    kind      = "ServiceAccount"
    name      = var.image_builder_k8s_service_account.name
    namespace = var.image_builder_k8s_service_account.namespace
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external-secrets-sa-trusted-cluster.namespace
    name      = var.external-secrets-sa-trusted-cluster.name
    api_group = "rbac.authorization.k8s.io"
  }
}
resource "kubernetes_cluster_role_binding" "access-signify-prod-secret-untrusted-workloads" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    name = "access-signify-secrets"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "access-signify-secrets"
  }
  subject {
    kind      = "ServiceAccount"
    name      = var.image_builder_k8s_service_account.name
    namespace = var.image_builder_k8s_service_account.namespace
  }
  subject {
    kind      = "ServiceAccount"
    namespace = var.external-secrets-sa-trusted-cluster.namespace
    name      = var.external-secrets-sa-trusted-cluster.name
    api_group = "rbac.authorization.k8s.io"
  }
}
