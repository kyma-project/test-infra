# This file deploys gatekeeper to the prow, workloads clusters.

module "prow_gatekeeper" {
  providers = {
    kubernetes = kubernetes.prow_k8s_cluster
    google     = google
    kubectl    = kubectl.prow_k8s_cluster
  }

  source = "../../modules/opa-gatekeeper"

  manifests_path = var.gatekeeper_manifest_path

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path          = [var.prow_constraints_path]
}

module "trusted_workload_gatekeeper" {
  providers = {
    kubernetes = kubernetes.trusted_workload_k8s_cluster
    google     = google
    kubectl    = kubectl.trusted_workload_k8s_cluster
  }
  source = "../../modules/opa-gatekeeper"

  manifests_path            = var.gatekeeper_manifest_path
  constraint_templates_path = [var.constraint_templates_path]
  constraints_path = [
    var.trusted_workloads_constraints_path,
    var.workloads_constraints_path
  ]
}

module "untrusted_workload_gatekeeper" {
  providers = {
    kubernetes = kubernetes.untrusted_workload_k8s_cluster
    google     = google
    kubectl    = kubectl.untrusted_workload_k8s_cluster
  }
  source = "../../modules/opa-gatekeeper"

  manifests_path = var.gatekeeper_manifest_path

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path = [
    var.untrusted_workloads_constraints_path,
    var.workloads_constraints_path
  ]
}
# (2025-03-04)