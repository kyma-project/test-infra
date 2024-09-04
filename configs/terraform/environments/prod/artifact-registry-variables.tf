###################################
# Artifact Registry related values
###################################
variable "kyma_project_artifact_registry_collection" {
  description = "Artifact Registry related data set"
  type = map(object({
    name  = string
    owner = string
    type  = string
    writer_serviceaccounts = optional(list(string), [])
    reader_serviceaccounts = list(string)
    primary_area = optional(string, "europe")
    multi_region = optional(bool, true)
    public = optional(bool, false)
    immutable = optional(bool, false)
  }))
}