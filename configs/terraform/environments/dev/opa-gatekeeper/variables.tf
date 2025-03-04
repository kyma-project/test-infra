variable "gatekeeper_manifest_path" {
  type    = string
  default = "../../../../../opa/gatekeeper/deployments/gatekeeper.yaml"
}

variable "managed_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  description = "Details of the k8s cluster to apply the manifest to."
}

variable "gcp_region" {
  type        = string
  default     = "europe-west4"
  description = "Default Google Cloud region to create resources."
}

variable "gcp_project_id" {
  type        = string
  default     = "sap-kyma-neighbors-dev"
  description = "Google Cloud project to create resources."
}

variable "constraint_templates_path" {
  type    = string
  default = "../../../../../opa/gatekeeper/constraint-templates/**.yaml"
}

variable "constraints_path" {
  type = string
}
# (2025-03-04)