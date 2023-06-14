# This file creates the terraform executor Google Cloud service account and k8s service account need to run terraform.
# k8s service account is created in the managed cluster.
# It grants required permissions to the Google Cloud service account and setup workload identity.

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

  external_secrets_sa = {
    name      = var.external_secrets_sa.name
    namespace = var.external_secrets_sa.namespace
  }
}
