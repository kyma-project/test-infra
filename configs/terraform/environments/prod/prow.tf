resource "google_container_cluster" "trusted_workload" {
  provider                 = google-beta
  name                     = "trusted-workload-kyma-prow"
  location                 = var.gcp_region
  remove_default_node_pool = true
  initial_node_count       = 1
  release_channel {
    channel = "REGULAR"
  }
  cluster_autoscaling {
    enabled             = false
    autoscaling_profile = "OPTIMIZE_UTILIZATION"
  }
  enable_shielded_nodes = true
  description           = "Prow control-plane cluster"
  workload_identity_config {
    workload_pool = "${var.gcp_project_id}.svc.id.goog"
  }
  resource_labels = {
    business_tag  = "corporate"
    exposure_tag  = "internet_ingress"
    landscape_tag = "production"
    name_cluster  = "trusted-workload-kyma-prow"
  }
}

resource "google_container_node_pool" "prowjobs_pool" {
  name    = "prowjobs-pool"
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
      workload = "prowjobs"
    }
  }
  management {
    auto_repair  = true
    auto_upgrade = true
  }
}

resource "google_container_node_pool" "components_pool" {
  cluster = google_container_cluster.trusted_workload.id
  name    = "components-pool"
  autoscaling {
    max_node_count  = 1
    min_node_count  = 1
    location_policy = "ANY"
  }
  node_config {
    workload_metadata_config {
      mode = "GKE_METADATA"
    }
    preemptible  = true
    machine_type = "n1-standard-2"
    metadata = {
      disable-legacy-endpoints = "true"
    }
    labels = {
      workload = "components"
    }
    taint {
      effect = "PREFER_NO_SCHEDULE"
      key    = "components.gke.io/gke-managed-components"
      value  = "true"
    }
  }
  management {
    auto_repair  = true
    auto_upgrade = true
  }
}