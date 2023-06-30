terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.55.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google" {
  alias = "workloads"
  project = var.workloads_project_id
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}
