output "terraform_executor_k8s_service_account" {
  value = kubernetes_service_account.terraform_executor
}
