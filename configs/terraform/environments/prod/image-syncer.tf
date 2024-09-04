resource "google_artifact_registry_repository_iam_member" "image_syncer_writer" {
  location   = google_artifact_registry_repository.prod_docker_repository.location
  repository = google_artifact_registry_repository.prod_docker_repository.name
  role       = "roles/artifactregistry.createOnPushWriter"
  member     = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_ref/${var.image_builder_reusable_workflow_ref}"
}