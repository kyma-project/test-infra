locals {
  artifact_registry_tags = {
    owner  = var.owner
    module = var.module
    type   = var.type
  }

  location = var.artifact_registry_multi_region == true ? var.artifact_registry_primary_area : var.gcp_region
}