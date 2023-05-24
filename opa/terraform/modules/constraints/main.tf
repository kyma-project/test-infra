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
  for_each = toset(flatten([
    for kpd in data.kubectl_path_documents.constraints_path : kpd.documents
  ]))
  yaml_body = each.value
}
