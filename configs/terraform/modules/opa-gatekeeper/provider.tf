terraform {
  required_version = ">= 1.8.0"

  required_providers {
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0.0"
    }
    google = {
      source  = "hashicorp/google"
      version = ">= 4.64.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.22.0"
    }
  }
}
