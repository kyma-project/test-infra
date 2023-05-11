module "tekton-gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path = var.tekton_gatekeeper_manifest_path
}
