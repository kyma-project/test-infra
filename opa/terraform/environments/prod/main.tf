module "tekton_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path     = var.tekton_gatekeeper_manifest_path
  k8s_config_path    = var.tekton_k8s_config_path
  k8s_config_context = var.tekton_k8s_config_context
}
module "trusted_workloads_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path     = var.trusted_workloads_gatekeeper_manifest_path
  k8s_config_path    = var.tekton_k8s_config_path
  k8s_config_context = var.tekton_k8s_config_context
}
module "untrusted_workloads_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path     = var.untrusted_workloads_gatekeeper_manifest_path
  k8s_config_path    = var.tekton_k8s_config_path
  k8s_config_context = var.tekton_k8s_config_context
}
