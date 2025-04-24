terraform {
  required_version = ">= 1.6.1"

  required_providers {

    google = {
      source  = "hashicorp/google"
      version = ">=4.76.0"
    }

  }
}
