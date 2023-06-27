variable "constraint_templates_path" {
  type    = string
  default = "../../../../opa/gatekeeper/constraint-templates/**.yaml"

  description = "Path to the constraint templates directory."
}

variable "tekton_constraints_path" {
  type    = string
  default = "../../../../tekton/deployments/gatekeeper-constraints/**.yaml"

  description = "Path to the tekton cluster constraints directory."
}

variable "prow_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/prow/**.yaml"

  description = "Path to the trusted workloads cluster constraints directory."
}

variable "trusted_workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/trusted/**.yaml"

  description = "Path to the trusted workloads cluster constraints directory."
}

variable "untrusted_workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/untrusted/**.yaml"

  description = "Path to the untrusted workloads cluster constraints directory."
}

variable "workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/workloads/**.yaml"

  description = "Path to both workload clusters constraints directory."
}
