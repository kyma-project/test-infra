module "artifact_registry" {
  artifact_registry_type         = var.artifact_registry_type
  artifact_registry_module       = var.artifact_registry_module
  artifact_registry_prefix       = var.artifact_registry_prefix
  immutable_artifact_registry    = var.immutable_artifact_registry
  artifact_registry_multi_region = var.artifact_registry_multi_region

  source = "../../../../artifact-registry/terraform/environments/prod"

}