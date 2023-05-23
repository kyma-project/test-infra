# This file deploys gatekeeper to the prow workloads and tekton clusters.

module "tekton_gatekeeper" {
  providers = {
    kubernetes = kubernetes.tekton_k8s_cluster
    google     = google
    kubectl    = kubectl.tekton_k8s_cluster
  }

  source = "../../../../opa/terraform/modules/gatekeeper"

  manifests_path = var.tekton_gatekeeper_manifest_path
}

module "trusted_workload_gatekeeper" {
  providers = {
    kubernetes = kubernetes.trusted_workload_k8s_cluster
    google     = google
    kubectl    = kubectl.trusted_workload_k8s_cluster
  }
  source = "../../../../opa/terraform/modules/gatekeeper"

  manifests_path = var.trusted_workload_gatekeeper_manifest_path
}

module "untrusted_workload_gatekeeper" {
  providers = {
    kubernetes = kubernetes.untrusted_workload_k8s_cluster
    google     = google
    kubectl    = kubectl.untrusted_workload_k8s_cluster
  }
  source = "../../../../opa/terraform/modules/gatekeeper"

  manifests_path = var.untrusted_workload_gatekeeper_manifest_path
}
