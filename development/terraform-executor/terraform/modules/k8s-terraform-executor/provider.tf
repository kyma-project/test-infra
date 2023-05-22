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

# Configure the Google Cloud client to configure the k8s provider.
# Client provides the access token to authenticate to the k8s cluster.
data "google_client_config" "gcp" {
}

# Get the k8s cluster details to configure the k8s provider.
# Cluster details provide the endpoint and cluster certificate to authenticate to the k8s cluster.
data "google_container_cluster" "managed_k8s_cluster" {
  name     = var.managed_k8s_cluster.name
  location = var.managed_k8s_cluster.location
}

provider "kubernetes" {
  host  = "https://${data.google_container_cluster.managed_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.managed_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}
