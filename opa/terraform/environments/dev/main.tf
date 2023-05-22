module "tekton_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path = var.gatekeeper_manifest_path
}
