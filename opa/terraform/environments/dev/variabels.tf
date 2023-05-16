variable "gatekeeper_manifest_path" {
  type    = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
}

variable "k8s_config_path" {
  type        = string
  default     = "~/.kube/config"
  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
}

variable "k8s_config_context" {
  type        = string
  description = "Context to use to connect to managed k8s cluster."
}
