module "test_gatekeeper" {
  source = "../../../modules/opa-gatekeeper"

  manifests_path = var.gatekeeper_manifest_path

  constraint_templates_path = [var.constraint_templates_path]
  constraints_path          = [var.constraints_path]
}
# (2025-03-04)