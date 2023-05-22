module "tekton_gatekeeper" {
  source = "../../modules/gatekeeper"

  manifests_path = var.gatekeeper_manifest_path
  #  managed_k8s_cluster = var.managed_k8s_cluster
  #
  #  gcp_region     = var.gcp_region
  #  gcp_project_id = var.gcp_project_id
}
