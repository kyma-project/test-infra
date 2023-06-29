locals {
  artifact_registry_tags = {
    owner  = var.artifact_registry_owner
    module = var.artifact_registry_module
    type   = var.artifact_registry_type
  }

  location = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : var.gcp_region
}