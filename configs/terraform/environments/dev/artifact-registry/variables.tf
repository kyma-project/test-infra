
###################################
# Artifact Registry related values
###################################
variable "artifact_registry_gcp_region" {
  type        = string
  description = "Default Google Cloud region to create resources."
}

variable "artifact_registry_gcp_project_id" {
  type        = string
  description = "Google Cloud project to create resources."
}

variable "artifact_registry_collection" {
  description = "Artifact Registry related data set"
  type = map(object({
    name                   = string
    owner                  = string
    type                   = string
    writer_serviceaccounts = optional(list(string), [])
    reader_serviceaccounts = list(string)
    primary_area           = optional(string, "europe")
    multi_region           = optional(bool, true)
    public                 = optional(bool, false)
    immutable              = optional(bool, false)
  }))
}# (2025-03-04)