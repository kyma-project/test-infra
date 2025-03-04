data "kubectl_file_documents" "automated_approver" {
  content = file(var.automated_approver_deployment_path)
}

resource "kubectl_manifest" "automated_approver" {
  provider  = kubectl.prow_k8s_cluster
  for_each  = data.kubectl_file_documents.automated_approver.manifests
  yaml_body = each.value
}

data "kubectl_file_documents" "automated_approver_rules" {
  content = file(var.automated_approver_rules_path)
}

resource "kubectl_manifest" "automated_approver_rules" {
  provider  = kubectl.prow_k8s_cluster
  for_each  = data.kubectl_file_documents.automated_approver_rules.manifests
  yaml_body = each.value
}
# (2025-03-04)