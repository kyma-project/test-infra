output "artifact_registry_list" {
  description = "Artifact Registry name"
  value       = google_artifact_registry_repository.artifact_registry[*]
}