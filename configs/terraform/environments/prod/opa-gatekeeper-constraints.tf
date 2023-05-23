module "tekton-gatekeeper-constraints" {
  providers = {
    kubectl = kubectl.tekton_k8s_cluster
  }

  source = "../../../../opa/terraform/modules/constraints"

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path          = [var.tekton_constraints_path]
}

module "untrusted_gatekeeper_constraints" {
  providers = {
    kubectl = kubectl.untrusted_workload_k8s_cluster
  }

  source = "../../../../opa/terraform/modules/constraints"

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path = [
    var.untrusted_workloads_constraints_path,
    var.workloads_constraints_path
  ]
}

module "trusted_gatekeeper_constraints" {
  providers = {
    kubectl = kubectl.trusted_workload_k8s_cluster
  }

  source = "../../../../opa/terraform/modules/constraints"

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path = [
    var.trusted_workloads_constraints_path,
    var.workloads_constraints_path
  ]
}
