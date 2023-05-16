terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-project"
    prefix = "prow"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.50.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.17.0"
    }
  }
}

variable "gcp_zone" {
  type = string
}

variable "gcp_region" {
  type = string
}

variable "gcp_project_id" {
  type = string
}

variable "k8s_config_path" {
  type        = string
  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
}

variable "k8s_config_context" {
  type        = string
  description = "Context to use to connect to managed k8s cluster."
}

provider "kubernetes" {
  config_path    = var.k8s_config_path
  config_context = var.k8s_config_context
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
  zone    = var.gcp_zone
}
