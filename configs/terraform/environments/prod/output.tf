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
  value = google_service_account.terraform_executor
}

output "terraform_executor_gcp_iam_member" {
  value = google_project_iam_member.terraform_executor_owner
}
output "terraform_executor_gcp_workload_identity" {
  value = google_service_account_iam_binding.terraform_workload_identity
}

output "trusted_workload_terraform_executor_k8s_service_account" {
  value = kubernetes_service_account.trusted_workload_terraform_executor
}

output "untrusted_workload_terraform_executor_k8s_service_account" {
  value = kubernetes_service_account.untrusted_workload_terraform_executor
}

output "artifact_registry_list" {
  description = "Artifact Registry name"
  value       = google_artifact_registry_repository.artifact_registry[*]
}