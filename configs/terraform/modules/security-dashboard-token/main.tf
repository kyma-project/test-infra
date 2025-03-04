terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.76.0"
    }
  }
}

variable "gcp_project_id" {
  type    = string
  default = "sap-kyma-prow"
}

# Used to retrieve project_number later
data "google_project" "project" {
  provider = google
}
# (2025-03-04)