output "secret_id" {
  value       = google_secret_manager_secret.secret.secret_id
  description = "The ID of the created Secret Manager secret."
}

output "secret_name" {
  value       = google_secret_manager_secret.secret.name
  description = "The full resource name of the created Secret Manager secret."
}
