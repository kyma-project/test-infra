terraform {
  required_providers {
    # google = {
    #   source  = "hashicorp/google"
    #   version = "4.55.0"
    # }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.7.0"
    }
  }
}

provider "kubectl" {
  config_path = "~/.kube/config"
}
