module "artifact_registry" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }


  for_each               = var.kyma_project_artifact_registry_collection
  registry_name          = each.value.name
  type                   = each.value.type
  immutable_tags         = each.value.immutable
  multi_region           = each.value.multi_region
  owner                  = each.value.owner
  writer_serviceaccounts = each.value.writer_serviceaccounts
  reader_serviceaccounts = each.value.reader_serviceaccounts
  public                 = each.value.public
}

resource "google_artifact_registry_repository_iam_member" "writer_service_account" {
  project    = data.google_client_config.this.project
  location   = var.multi_region == true ? var.primary_area : data.google_client_config.this.region
  repository = google_artifact_registry_repository.artifact_registry.name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${each.value}"
}