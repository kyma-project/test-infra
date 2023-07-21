module "tekton_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path = var.gatekeeper_manifest_path

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path          = [var.tekton_constraints_path]
}
