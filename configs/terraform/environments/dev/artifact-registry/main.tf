module "artifact_registry" {
  source = "../../../modules/artifact-registry"

  for_each                         = toset(var.artifact_registry_names)
  artifact_registry_name           = each.value
  gcp_region                       = var.gcp_region
  artifact_registry_type           = var.artifact_registry_type
  artifact_registry_module         = var.artifact_registry_module
  immutable_artifact_registry      = var.immutable_artifact_registry
  artifact_registry_multi_region   = var.artifact_registry_multi_region
  artifact_registry_owner          = var.artifact_registry_owner
  artifact_registry_serviceaccount = var.artifact_registry_serviceaccount


}