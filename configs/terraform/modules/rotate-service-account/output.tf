output "service_account_keys_rotator_service_account" {
  value = google_service_account.service_account_keys_rotator
}

output "service_account_keys_rotator_service_account_iam" {
  value = google_project_iam_member.service_account_keys_rotator
}

output "service_account_keys_rotator_cloud_run_service" {
  value = {
    id       = google_cloud_run_service.service_account_keys_rotator.id
    name     = google_cloud_run_service.service_account_keys_rotator.name
    location = google_cloud_run_service.service_account_keys_rotator.location
    project  = google_cloud_run_service.service_account_keys_rotator.project
    status   = google_cloud_run_service.service_account_keys_rotator.status
  }
}

output "service_account_keys_rotator_subscription" {
  value = google_pubsub_subscription.service_account_keys_rotator
}
