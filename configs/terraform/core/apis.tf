terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-project"
    prefix = "core"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.0.1"
    }
  }
}

variable "gcp_project_id" {
  type = string
}

# Used to retrieve project_number later
data "google_project" "project" {
  provider   = google
  project_id = var.gcp_project_id
}

# Enable Eventarc API
resource "google_project_service" "eventarc" {
  provider           = google
  service            = "eventarc.googleapis.com"
  project            = data.google_project.project.number
  disable_on_destroy = false
}

# Enable Pub/Sub API
resource "google_project_service" "pubsub" {
  provider           = google
  service            = "pubsub.googleapis.com"
  project            = data.google_project.project.number
  disable_on_destroy = false
}

# Enable Workflows API
resource "google_project_service" "workflows" {
  provider           = google
  service            = "workflows.googleapis.com"
  project            = data.google_project.project.number
  disable_on_destroy = false
}
