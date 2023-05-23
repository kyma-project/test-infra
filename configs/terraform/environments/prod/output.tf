output "tekton_gatekeeper" {
  value     = module.tekton_gatekeeper
  sensitive = true
}

output "trusted_workload_gatekeeper" {
  value     = module.trusted_workload_gatekeeper
  sensitive = true
}

output "untrusted_workload_gatekeeper" {
  value     = module.untrusted_workload_gatekeeper
  sensitive = true
}

output "terraform_executor_gcp_service_account" {
  value     = module.terraform_executor_gcp_service_account
  sensitive = true
}

output "tekton_terraform_executor_k8s_service_account" {
  value = module.tekton_terraform_executor_k8s_service_account
}

output "trusted_workload_terraform_executor_k8s_service_account" {
  value = module.trusted_workload_terraform_executor_k8s_service_account
}

output "untrusted_workload_terraform_executor_k8s_service_account" {
  value = module.untrusted_workload_terraform_executor_k8s_service_account
}


output "tekton_gatekeeper_constraints" {
  value     = module.tekton_gatekeeper_constraints
  sensitive = true
}

output "trusted_gatekeeper_constraints" {
  value     = module.trusted_gatekeeper_constraints
  sensitive = true
}

output "untrusted_gatekeeper_constraints" {
  value     = module.untrusted_gatekeeper_constraints
  sensitive = true
}
