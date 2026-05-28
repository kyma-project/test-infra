output "signify_secret_rotator_service_account" {
  value = google_service_account.signify_secret_rotator
}

output "signify_secret_rotator_cloud_run_service" {
  value = {
    id       = google_cloud_run_service.signify_secret_rotator.id
    name     = google_cloud_run_service.signify_secret_rotator.name
    location = google_cloud_run_service.signify_secret_rotator.location
    project  = google_cloud_run_service.signify_secret_rotator.project
    status   = google_cloud_run_service.signify_secret_rotator.status
  }
}

output "signify_secret_rotator_subscription" {
  value = google_pubsub_subscription.signify_secret_rotator
}
