output "service_account_keys_cleaner_service_account" {
  value = google_service_account.service_account_keys_cleaner
}

output "service_account_keys_cleaner_cloud_run_service" {
  value = {
    id       = google_cloud_run_service.service_account_keys_cleaner.id
    name     = google_cloud_run_service.service_account_keys_cleaner.name
    location = google_cloud_run_service.service_account_keys_cleaner.location
    project  = google_cloud_run_service.service_account_keys_cleaner.project
    status   = google_cloud_run_service.service_account_keys_cleaner.status
  }
}

output "service_account_keys_cleaner_secheduler" {
  value = google_cloud_scheduler_job.service_account_keys_cleaner
}
