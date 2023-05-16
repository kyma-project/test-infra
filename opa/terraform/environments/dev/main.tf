module "tekton_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path = var.gatekeeper_manifest_path

  k8s_config_path    = var.k8s_config_path
  k8s_config_context = var.k8s_config_context
}
