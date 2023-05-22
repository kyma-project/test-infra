terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.64.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.20.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.14.0"
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# Configure the Google Cloud client to configure the k8s provider.
# Client provides the access token to authenticate to the k8s cluster.
data "google_client_config" "gcp" {
}

# Get the k8s cluster details to configure the k8s provider.
# Cluster details provide the endpoint and cluster certificate to authenticate to the k8s cluster.
data "google_container_cluster" "tekton_k8s_cluster" {
  name     = var.tekton_k8s_cluster.name
  location = var.tekton_k8s_cluster.location
}

provider "kubernetes" {
  alias = "tekton_k8s_cluster"
  host  = "https://${data.google_container_cluster.tekton_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.tekton_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
}

provider "kubectl" {
  alias = "tekton_k8s_cluster"
  host  = "https://${data.google_container_cluster.tekton_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.tekton_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
  load_config_file = false
}

# Get the k8s cluster details to configure the k8s provider.
# Cluster details provide the endpoint and cluster certificate to authenticate to the k8s cluster.
data "google_container_cluster" "trusted_workload_k8s_cluster" {
  name     = var.trusted_workload_k8s_cluster.name
  location = var.trusted_workload_k8s_cluster.location
}

provider "kubernetes" {
  alias = "trusted_workload_k8s_cluster"
  host  = "https://${data.google_container_cluster.trusted_workload_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.trusted_workload_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
}

provider "kubectl" {
  alias = "trusted_workload_k8s_cluster"
  host  = "https://${data.google_container_cluster.trusted_workload_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.trusted_workload_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
  load_config_file = false
}

# Get the k8s cluster details to configure the k8s provider.
# Cluster details provide the endpoint and cluster certificate to authenticate to the k8s cluster.
data "google_container_cluster" "untrusted_workload_k8s_cluster" {
  name     = var.untrusted_workload_k8s_cluster.name
  location = var.untrusted_workload_k8s_cluster.location
}

provider "kubernetes" {
  alias = "untrusted_workload_k8s_cluster"
  host  = "https://${data.google_container_cluster.untrusted_workload_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.untrusted_workload_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
}

provider "kubectl" {
  alias = "untrusted_workload_k8s_cluster"
  host  = "https://${data.google_container_cluster.untrusted_workload_k8s_cluster.endpoint}"
  token = data.google_client_config.gcp.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.untrusted_workload_k8s_cluster.master_auth[0].cluster_ca_certificate,
  )
  load_config_file = false
}
