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
