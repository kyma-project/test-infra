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

data "kubectl_path_documents" "constraint_templates_path" {
  for_each = toset(var.constraint_templates_path)
  pattern  = each.value
}

resource "kubectl_manifest" "constraint_templates" {
  for_each = toset(flatten([
    for kpd in data.kubectl_path_documents.constraint_templates_path : kpd.documents
  ]))
  yaml_body = each.value
}


data "kubectl_path_documents" "constraints_path" {
  for_each = toset(var.constraints_path)
  pattern  = each.value
}

resource "kubectl_manifest" "constraints" {
  depends_on = [kubectl_manifest.constraint_templates]
  for_each = toset(flatten([
    for kpd in data.kubectl_path_documents.constraints_path : kpd.documents
  ]))
  yaml_body = each.value
}
# (2025-03-04)