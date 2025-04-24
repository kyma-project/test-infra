terraform {
  required_version = ">= 1.8.0"

  required_providers {

    google = {
      source  = "hashicorp/google"
      version = ">=4.76.0"
    }

  }
}
