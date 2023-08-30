provider "google" {
  alias   = "artifact_registry_kyma_project"
  project = var.artifact_registry_gcp_project_id_kyma_project
  region  = var.artifact_registry_gcp_region_kyma_project
}

module "artifact_registry" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.artifact_registry_kyma_project
  }


  for_each               = var.artifact_registry_collection_kyma_project
  registry_name          = each.value.name
  type                   = each.value.type
  immutable_tags         = each.value.immutable
  multi_region           = each.value.multi_region
  owner                  = each.value.owner
  writer_serviceaccounts = each.value.writer_serviceaccounts
  reader_serviceaccounts = each.value.reader_serviceaccounts
  public                 = each.value.public
}