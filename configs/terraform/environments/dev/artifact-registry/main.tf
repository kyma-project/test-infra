
module "artifact_registry" {
  source = "../../../modules/artifact-registry"

  providers = {
    google = google.artifact_registry_smart_tractor
  }

  for_each               = var.artifact_registry_collection
  registry_name          = each.value.name
  type                   = each.value.type
  immutable_tags         = each.value.immutable
  multi_region           = each.value.multi_region
  owner                  = each.value.owner
  writer_serviceaccounts = each.value.writer_serviceaccounts
  reader_serviceaccounts = each.value.reader_serviceaccounts
  public                 = each.value.public
}# (2025-03-04)