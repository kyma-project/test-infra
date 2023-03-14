output "service_account_keys_rotator_service_account" {
  value = google_service_account.service_account_keys_rotator
}

output "service_account_keys_rotator_service_account_iam" {
  value = google_project_iam_member.service_account_keys_rotator
}

output "service_account_keys_rotator_cloud_run_service" {
  value = google_cloud_run_service.service_account_keys_rotator
}

output "service_account_keys_rotator_subscription" {
  value = google_pubsub_subscription.service_account_keys_rotator
}
