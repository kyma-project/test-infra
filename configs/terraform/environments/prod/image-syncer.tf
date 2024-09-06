resource "google_artifact_registry_repository_iam_member" "image_syncer_prod_repo_writer" {
  location   = google_artifact_registry_repository.prod_docker_repository.location
  repository = google_artifact_registry_repository.prod_docker_repository.name
  role       = "roles/artifactregistry.createOnPushWriter"
  member = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.reusable_workflow_run/event_name:push:repository_owner_id:${data.github_organization.kyma-project.id}:reusable_workflow_ref:${var.image_syncer_reusable_workflow_ref}"
}