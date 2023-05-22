variable "manifests_path" {
  type        = string
  description = "Path to the gatekeeper manifest file to apply to the cluster."
}

#variable "managed_k8s_cluster" {
#  type = object({
#    name     = string
#    location = string
#  })
#
#  description = "Details of the k8s cluster to apply the manifest to."
#}
#
#variable "gcp_region" {
#  type = string
#  description = "Default Google Cloud region to create resources."
#}
#
#variable "gcp_project_id" {
#  type = string
#  description = "Google Cloud project to create resources."
#}
#
