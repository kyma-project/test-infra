terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.64.0"
    }
  }
}

provider "google" {
  alias   = "artifact_registry_smart_tractor"
  project = var.artifact_registry_gcp_project_id
  region  = var.artifact_registry_gcp_region
}

# (2025-03-04)