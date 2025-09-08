terraform {
  required_version = ">= 1.6.1"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.0.1"
    }

    google-beta = {
      source  = "hashicorp/google-beta"
      version = ">= 6.0.1"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.22.0"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.6.0"
    }
  }
}

# data.google_client_config configures Google Cloud client.
# Google Cloud client provides the access token to authenticate to the k8s cluster.
# Access token is used to configure the k8s and kubectl providers.
# data.google_container_cluster gets the k8s cluster details.
# Cluster details provides the endpoint and cluster certificate to authenticate to the k8s cluster.
# Cluster details are used to configure the k8s and kubectl providers.

provider "github" {
  alias = "kyma_project"
  owner = var.kyma-project-github-org
}

# sap-kyma-prow project provider
provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

provider "google" {
  alias   = "kyma_project"
  project = var.kyma_project_gcp_project_id
  region  = var.kyma_project_gcp_region
}

provider "google-beta" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

data "google_client_config" "gcp" {
}
