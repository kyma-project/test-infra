###################################
# Artifact Registry related values
###################################
variable "kyma_project_artifact_registry_collection" {
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
}

variable "prod_docker_repository" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    mode                   = string
    cleanup_policy_dry_run = bool
    labels = map(string)
  })
  default = {
    name                   = "prod"
    description            = "Production images for kyma-project"
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    mode                   = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run = true
    labels = {
      "type" = "production"
    }
  }
}