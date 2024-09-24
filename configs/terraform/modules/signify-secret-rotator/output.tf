output "signify_secret_rotator_service_account" {
  value = google_service_account.signify_secret_rotator
}

output "signify_secret_rotator_cloud_run_service" {
  value = google_cloud_run_service.signify_secret_rotator
}

output "signify_secret_rotator_subscription" {
  value = google_pubsub_subscription.signify_secret_rotator
}
