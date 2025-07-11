output "terraform_executor_gcp_service_account" {
  value = google_service_account.terraform_executor
}

output "terraform_executor_gcp_prow_project_iam_member" {
  value = google_project_iam_member.terraform_executor_prow_project_owner
}

output "terraform_executor_gcp_workload_identity" {
  value = google_service_account_iam_binding.terraform_workload_identity
}

output "artifact_registry" {
  value     = module.artifact_registry
  sensitive = false
}

output "secrets_rotator_dead_letter_topic" {
  value = google_pubsub_topic.secrets_rotator_dead_letter
}

output "secrets-rotator" {
  value = google_service_account.secrets-rotator
}

output "secret-manager-notifications-topic" {
  value = data.google_pubsub_topic.secret-manager-notifications-topic
}