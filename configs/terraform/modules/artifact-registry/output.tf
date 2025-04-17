output "artifact_registry_collection" {
  description = "Artifact Registry"
  value       = google_artifact_registry_repository.artifact_registry
}

output "location" {
  description = "Location of the Artifact Registry"
  value       = google_artifact_registry_repository.artifact_registry.location
}

output "repository_id" {
  description = "ID of the Artifact Registry"
  value       = google_artifact_registry_repository.artifact_registry.repository_id
}
