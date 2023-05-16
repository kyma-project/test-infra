module "tekton-gatekeeper-constraints" {
  source = "../../modules/constraints"

  # manifests_path = var.tekton_gatekeeper_manifest_path
  constraint_templates_path = var.constraint_templates_path
  constraints_path          = var.tekton_constraints_path
}
