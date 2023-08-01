terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.64.0"
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}
