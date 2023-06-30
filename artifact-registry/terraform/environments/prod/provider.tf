terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.64.0"
    }
  }
}

# data.google_client_config configures Google Cloud client.
# Google Cloud client provides the access token to authenticate to the k8s cluster.
# Access token is used to configure the k8s and kubectl providers.
# data.google_container_cluster gets the k8s cluster details.
# Cluster details provides the endpoint and cluster certificate to authenticate to the k8s cluster.
# Cluster details are used to configure the k8s and kubectl providers.

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}
