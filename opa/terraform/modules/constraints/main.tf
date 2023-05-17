resource "kubectl_manifest" "constraint_templates" {
  for_each = flatten([
    for constraint_templates_path in var.constraint_templates_path : fileset(constraint_templates_path, "**.yaml")
  ])
  yaml_body = each.value
  # wait_for_rollout = false
}


resource "kubectl_manifest" "constraints" {
  for_each = flatten([
    for constraints_path in var.constraints_path : fileset(constraints_path, "**.yaml")
  ]) #fileset(var.constraints_path, "**.yaml")
  yaml_body = each.value
  # wait_for_rollout = false

}
