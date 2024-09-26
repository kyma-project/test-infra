terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.55.0"
    }

    google-beta = {
      source = "hashicorp/google-beta"
    }
  }
}