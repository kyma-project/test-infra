module "tekton_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path = var.gatekeeper_manifest_path
}

module "tekton_gatekeeper_constraints" {
  providers = {
    kubectl = kubectl
  }

  source = "../../modules/constraints"

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path          = [var.tekton_constraints_path]
}
