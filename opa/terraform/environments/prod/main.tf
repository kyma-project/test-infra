module "tekton-gatekeeper-constraints" {
  source = "../../modules/constraints"

  # manifests_path = var.tekton_gatekeeper_manifest_path
  constraint_templates_path = var.constraint_templates_path
  constraints_path          = var.tekton_constraints_path
}

module "untrusted-gatekeeper-constraints" {
  source = "../../modules/constraints"

  # manifests_path = var.tekton_gatekeeper_manifest_path
  constraint_templates_path = var.constraint_templates_path
  constraints_path          = var.untrusted_workloads_constraints_path
}
module "untrusted-gatekeeper-constraints-common" {
  source = "../../modules/constraints"

  # manifests_path = var.tekton_gatekeeper_manifest_path
  constraint_templates_path = var.constraint_templates_path
  constraints_path          = var.workloads_constraints_path
}

module "trusted-gatekeeper-constraints" {
  source = "../../modules/constraints"

  # manifests_path = var.tekton_gatekeeper_manifest_path
  constraint_templates_path = var.constraint_templates_path
  constraints_path          = var.trusted_workloads_constraints_path
}
module "trusted-gatekeeper-constraints-common" {
  source = "../../modules/constraints"

  # manifests_path = var.tekton_gatekeeper_manifest_path
  constraint_templates_path = var.constraint_templates_path
  constraints_path          = var.workloads_constraints_path
}
