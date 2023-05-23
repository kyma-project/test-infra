variable "constraint_templates_path" {
  type    = string
  default = "../../../gatekeeper/constraint-templates"
}

variable "tekton_constraints_path" {
  type    = string
  default = "../../../../tekton/deployments/gatekeeper-constraints"
}

variable "k8s_config_path" {
  type        = string
  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
}

variable "k8s_config_context" {
  type        = string
  description = "Context to use to connect to managed k8s cluster."
}
