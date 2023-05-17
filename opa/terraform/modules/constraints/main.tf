resource "kubectl_manifest" "constraint_templates" {
  for_each  = fileset(var.constraint_templates_path, "**.yaml")
  yaml_body = each.value
  # wait_for_rollout = false
}

resource "kubectl_manifest" "constraints" {
  for_each  = fileset(var.constraints_path, "**.yaml")
  yaml_body = each.value
  # wait_for_rollout = false
}
