# This file deploys gatekeeper to the prow workloads and tekton clusters.

module "tekton_gatekeeper" {
  providers = {
    kubernetes = kubernetes.tekton_k8s_cluster
    google     = google
    kubectl    = kubectl.tekton_k8s_cluster
  }

  source = "../../../../opa/terraform/modules/gatekeeper"

  manifests_path = var.tekton_gatekeeper_manifest_path
  #  managed_k8s_cluster = var.tekton_k8s_cluster
  #
  #  gcp_region     = var.gcp_region
  #  gcp_project_id = var.gcp_project_id
}

module "trusted_workload_gatekeeper" {
  providers = {
    kubernetes = kubernetes.trusted_workload_k8s_cluster
    google     = google
    kubectl    = kubectl.trusted_workload_k8s_cluster
  }
  source = "../../../../opa/terraform/modules/gatekeeper"

  manifests_path = var.trusted_workload_gatekeeper_manifest_path
  #  managed_k8s_cluster = var.trusted_workload_k8s_cluster
  #
  #  gcp_region     = var.gcp_region
  #  gcp_project_id = var.gcp_project_id
}

module "untrusted_workload_gatekeeper" {
  providers = {
    kubernetes = kubernetes.untrusted_workload_k8s_cluster
    google     = google
    kubectl    = kubectl.untrusted_workload_k8s_cluster
  }
  source = "../../../../opa/terraform/modules/gatekeeper"

  manifests_path = var.untrusted_workload_gatekeeper_manifest_path
  #  managed_k8s_cluster = var.untrusted_workload_k8s_cluster
  #
  #  gcp_region     = var.gcp_region
  #  gcp_project_id = var.gcp_project_id
}
