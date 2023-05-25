# Read the Gatekeeper yaml documents from manifest file.
data "kubectl_file_documents" "gatekeeper" {
  content = file(var.manifests_path)
}

# Apply the Gatekeeper yaml documents.
# Do not wait for the Gatekeeper pods to be ready as the Gatekeeper pods will not be ready until some resource quotas are not applied.
resource "kubectl_manifest" "gatekeeper" {
  for_each         = data.kubectl_file_documents.gatekeeper.manifests
  yaml_body        = each.value
  wait_for_rollout = false
}
