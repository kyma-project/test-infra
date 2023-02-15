variable "prow_terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })
  description = "Details of terraform k8s service account."
}

resource "kubernetes_secret" "prow_terraform_executor" {
  metadata {
    name      = var.prow_terraform_executor_k8s_service_account.name
    namespace = var.prow_terraform_executor_k8s_service_account.namespace
    annotations = {
      "kubernetes.io/service-account.name" = kubernetes_service_account.prow_terraform_executor.metadata[0].name
    }
  }
  type = "kubernetes.io/service-account-token"
}


resource "kubernetes_service_account" "prow_terraform_executor" {
  metadata {
    namespace = var.prow_terraform_executor_k8s_service_account.namespace
    name      = var.prow_terraform_executor_k8s_service_account.name
    annotations = {
      "iam.gke.io/gcp-service-account" = format("%s@%s.iam.gserviceaccount.com", var.prow_terraform_executor_gcp_service_account.id, var.gcp_project_id)
    }
  }
  automount_service_account_token = true
}
