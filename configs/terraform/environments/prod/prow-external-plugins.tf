data "kubectl_file_documents" "automated_approver" {
  content = file(var.automated_approver_deployment_path)
}

resource "kubectl_manifest" "automated_approver" {
  for_each  = data.kubectl_file_documents.automated_approver.manifests
  yaml_body = each.value
}
