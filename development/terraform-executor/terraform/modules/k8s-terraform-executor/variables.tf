variable "terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })
  description = "Details of terraform executor k8s service account."
}

variable "terraform_executor_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })
  description = "Details of terraform executor gcp service account."
}

#variable "managed_k8s_cluster" {
#  type = object({
#    name     = string
#    location = string
#  })
#  description = "Details of the managed k8s cluster to apply the manifest to."
#}

#variable "gcp_region" {
#  type        = string
#  description = "Default Google Cloud region to create resources."
#}
#
#variable "gcp_project_id" {
#  type        = string
#  description = "Google Cloud project to create resources."
#}
