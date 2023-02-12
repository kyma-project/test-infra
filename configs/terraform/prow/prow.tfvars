prow_terraform_executor_gcp_service_account = {
  id = "prow-terraform-executor"
}

prow_terraform_executor_k8s_service_account = {
  name      = "prow-terraform-executor"
  namespace = "default"
}
