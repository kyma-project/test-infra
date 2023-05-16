variable "manifests_path" {
  type        = string
  description = "Path to the manifest file to apply to the cluster."
}

variable "k8s_config_path" {
  type        = string
  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
}

variable "k8s_config_context" {
  type        = string
  description = "Context to use to connect to managed k8s cluster."
}
