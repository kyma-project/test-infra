module "artifact_registry" {
  gcp_project_id                 = var.gcp_project_id
  gcp_region                     = var.gcp_region
  artifact_registry_type         = var.artifact_registry_type
  artifact_registry_module       = var.artifact_registry_module
  immutable_artifact_registry    = var.immutable_artifact_registry
  artifact_registry_multi_region = var.artifact_registry_multi_region

  source = "../../../../configs/terraform/modules/artifact-registry"

}

