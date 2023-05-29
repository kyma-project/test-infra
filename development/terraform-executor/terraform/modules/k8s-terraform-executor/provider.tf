terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.20.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "4.64.0"
    }
  }
}
