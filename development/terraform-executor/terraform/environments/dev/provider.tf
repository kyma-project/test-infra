terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.64.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.20.0"
    }
  }
}

provider "kubernetes" {
  config_path    = var.k8s_config_path
  config_context = var.k8s_config_context
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}
