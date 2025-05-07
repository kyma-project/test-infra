# TODO (dekiel): remove after migration to modulectl is done
variable "kyma_project_artifact_registry_collection" {
  type = map(object({
    name                      = string
    owner                     = string
    type                      = string
    description               = string
    repoAdmin_serviceaccounts = optional(list(string), [])
    writer_serviceaccounts    = optional(list(string), [])
    reader_serviceaccounts    = optional(list(string), [])
    primary_area              = optional(string, "europe")
    multi_region              = optional(bool, true)
    public                    = optional(bool, false)
    immutable                 = optional(bool, false)
    cleanup_policy_dry_run    = optional(bool, false)
    cleanup_policies = optional(list(object({
      id     = string
      action = string
      condition = optional(object({
        tag_state             = string
        tag_prefixes          = optional(list(string), [])
        package_name_prefixes = optional(list(string), [])
        older_than            = optional(string, "")
      }))
    })))
  }))
}


variable "prod_docker_repository" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    type                   = string
    cleanup_policy_dry_run = bool
    cleanup_policies = optional(list(object({
      id     = string
      action = string
      condition = optional(object({
        tag_state             = string
        tag_prefixes          = optional(list(string), [])
        package_name_prefixes = optional(list(string), [])
        older_than            = optional(string, "")
      }))
    })))
  })
  default = {
    name                   = "prod"
    description            = "Production images for kyma-project"
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    type                   = "production"
    cleanup_policy_dry_run = false
    cleanup_policies = [
      {
        id     = "delete-untagged"
        action = "DELETE"
        condition = {
          tag_state = "UNTAGGED"
        }
      }
    ]
  }
}

variable "docker_dev_repository" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    mode                   = string
    type                   = string
    cleanup_policy_dry_run = bool
    cleanup_policies = optional(list(object({
      id     = string
      action = string
      condition = optional(object({
        tag_state             = string
        tag_prefixes          = optional(list(string), [])
        package_name_prefixes = optional(list(string), [])
        older_than            = optional(string, "")
      }))
    })))
  })
  default = {
    name                   = "dev"
    description            = "Development images for kyma-project"
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    mode                   = "STANDARD_REPOSITORY"
    type                   = "development"
    cleanup_policy_dry_run = false
    cleanup_policies = [{
      id     = "delete-untagged"
      action = "DELETE"
      condition = {
        tag_state = "UNTAGGED"
      }
      },
      {
        id     = "delete-old-pr-images"
        action = "DELETE"
        condition = {
          tag_state = "TAGGED"
          # Google provider does not support the time units,
          # so we need to provide the time in seconds.
          # Time after which the images will be deleted.
          older_than   = "2592000s" # 2592000s = 720h = 30 days
          tag_prefixes = ["PR-"]
        }
    }]

  }
}
