module "artifact_registry" {
  source = "../../../modules/artifact-registry"

  for_each                                 = var.artifact_registry_collection
  artifact_registry_name                   = each.value.name
  artifact_registry_type                   = each.value.type
  immutable_artifact_registry              = each.value.immutable
  artifact_registry_multi_region           = each.value.multi_region
  artifact_registry_owner                  = each.value.owner
  artifact_registry_writer_serviceaccount  = each.value.writer_serviceaccount
  artifact_registry_reader_serviceaccounts = each.value.reader_serviceaccounts
  artifact_registry_public                 = each.value.public
}