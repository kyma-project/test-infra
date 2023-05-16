variable "constraint_templates_path" {
  type    = string
  default = "../../../gatekeeper/constraint-templates"
}

variable "var.tekton_constraints_path" {
  type    = string
  default = "../../../../tekton/deployments/gatekeeper-constraints"
}

variable "var.trusted_workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/trusted"
}

variable "var.untrusted_workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/untrusted"
}

variable "var.workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/workloads"
}
