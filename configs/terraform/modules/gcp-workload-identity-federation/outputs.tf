output "provider_name" {
  description = "Workload identity federation provider name"
  value       = google_iam_workload_identity_pool_provider.main.name
}
