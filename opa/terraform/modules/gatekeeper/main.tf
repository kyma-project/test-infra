data "kubectl_file_documents" "gatekeeper" {
  content = file(var.manifests_path)
}

resource "kubectl_manifest" "gatekeeper" {
  for_each         = data.kubectl_file_documents.gatekeeper.manifests
  yaml_body        = each.value
  wait_for_rollout = false
}
