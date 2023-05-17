variable "gatekeeper_manifest_path" {
  type    = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
}

variable "managed_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  description = "Name of the managed k8s cluster to apply the manifest to."
}

variable "gcp_region" {
  type    = string
  default = "europe-west4"
}

variable "gcp_project_id" {
  type    = string
  default = "sap-kyma-neighbors-dev"
}

#variable "k8s_config_path" {
#  type        = string
#  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
#}
#
#variable "k8s_config_context" {
#  type        = string
#  description = "Context to use to connect to managed k8s cluster."
#}
