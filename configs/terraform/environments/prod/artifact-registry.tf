provider "google" {
  alias   = "artifact_registry_kyma_project"
  project = var.artifact_registry_gcp_project_id
  region  = var.artifact_registry_gcp_region
}

module "artifact_registry" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.artifact_registry_kyma_project
  }


  for_each                                 = var.artifact_registry_collection
  artifact_registry_name                   = each.value.name
  artifact_registry_type                   = each.value.type
  artifact_registry_immutable_tags         = each.value.immutable
  artifact_registry_multi_region           = each.value.multi_region
  artifact_registry_owner                  = each.value.owner
  artifact_registry_writer_serviceaccounts = each.value.writer_serviceaccounts
  artifact_registry_reader_serviceaccounts = each.value.reader_serviceaccounts
  artifact_registry_public                 = each.value.public
}