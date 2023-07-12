resource "google_container_cluster" "trusted_workload" {
  name                     = "trusted-workload-kyma-prow"
  location                 = var.gcp_region
  remove_default_node_pool = true
  initial_node_count       = 1
  release_channel {
    channel = "REGULAR"
  }
  enable_shielded_nodes = true
  description           = "Prow control-plane cluster"
  workload_identity_config {
    workload_pool = "${var.gcp_project_id}.svc.id.goog"
  }
  resource_labels = {
    business_tag = "corporate"
    exposure_tag = "internet_ingress"
    landscape_tag = "production"
    name_cluster = "trusted-workload-kyma-prow"
  }
}

resource "google_container_node_pool" "preemptible_standard_pool" {
  name    = "standard-pool"
  cluster = google_container_cluster.trusted_workload.id
  autoscaling {
    max_node_count  = 16
    min_node_count  = 0
    location_policy = "ANY"
  }
  node_config {
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    preemptible  = true
    machine_type = "n2d-standard-8"
    disk_size_gb = 200
    metadata = {
      disable-legacy-endpoints = "true"
    }
    labels = {
      workload = "prow-jobs"
    }
  }
}