module "terraform_executor_gcp_service_account" {
  source = "../../modules/gcp-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }
}

module "terraform_executor_k8s_service_account" {
  source = "../../modules/k8s-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }
  k8s_config_path    = var.k8s_config_path
  k8s_config_context = var.k8s_config_context
}
