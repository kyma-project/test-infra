output "terraform_executor_gcp_service_account" {
  value = module.terraform_executor_gcp_service_account
}

output "tekton_terraform_executor_k8s_service_account" {
  value = module.tekton_terraform_executor_k8s_service_account
}

output "trusted_terraform_executor_k8s_service_account" {
  value = module.trusted_workloads_terraform_executor_k8s_service_account
}

output "untrusted_terraform_executor_k8s_service_account" {
  value = module.untrusted_workloads_terraform_executor_k8s_service_account
}
