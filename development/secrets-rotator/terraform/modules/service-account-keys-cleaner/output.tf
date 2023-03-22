output "service_account_keys_cleaner_service_account" {
  value = google_service_account.service_account_keys_cleaner
}

output "service_account_keys_cleaner_cloud_run_service" {
  value = google_cloud_run_service.service_account_keys_cleaner
}

output "service_account_keys_cleaner_secheduler" {
  value = google_cloud_scheduler_job.service_account_keys_cleaner
}
