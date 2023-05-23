output "terraform_executor_gcp_service_account" {
  value = google_service_account.terraform_executor
}

output "terraform_executor_workload_identity" {
  value = google_service_account_iam_binding.terraform_workload_identity
}
