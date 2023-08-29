
module "artifact_registry" {
  source = "../../../modules/artifact-registry"

  providers = {
    google = google.artifact_registry_smart_tractor
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