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

# TODO (dekiel): move to the module modules/artifact-registry
variable "docker_dev_repository" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    mode                   = string
    cleanup_policy_dry_run = bool
    pr_images_max_age      = string
    pr_images_tag_prefix   = string
    labels                 = map(string)
  })
  default = {
    name                   = "dev"
    description            = "Development images for kyma-project"
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    mode                   = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run = false
    # Google provider does not support the time units,
    # so we need to provide the time in seconds.
    # Time after which the images will be deleted.
    pr_images_max_age    = "2592000s" # 2592000s = 720h = 30 days
    pr_images_tag_prefix = "PR-"
    labels = {
      "type" = "development"
    }
  }
}
