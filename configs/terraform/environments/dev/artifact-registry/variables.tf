
###################################
# Artifact Registry related values
###################################
variable "gcp_region" {
  type        = string
  description = "Default Google Cloud region to create resources."
}

variable "gcp_project_id" {
  type        = string
  description = "Google Cloud project to create resources."
}

variable "artifact_registry_list" {
  description = "Artifact Registry related data set"
  type = list(object({
    name                   = string
    owner                  = string
    type                   = string
    writer_serviceaccount  = optional(string, "")
    reader_serviceaccounts = list(string)
    primary_area           = optional(string, "europe")
    multi_region           = optional(bool, true)
    public                 = optional(bool, false)
    immutable              = optional(bool, false)
  }))
}

