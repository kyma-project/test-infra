variable "constraint_templates_path" {
  type    = string
  default = "../../../gatekeeper/constraint-templates"
}

variable "var.tekton_constraints_path" {
  type    = string
  default = "../../../../tekton/deployments/gatekeeper-constraints"
}
