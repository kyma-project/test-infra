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
    cleanup_policies       = list(object({
      id      = string
      action  = string
      condition = object({
        tag_state = string
        older_than = string
      })
    }))
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
    cleanup_policies = [ {
      id = "delete_untagged"
      action = "DELETE"
      condition = {
        tag_state = "UNTAGGED"
        older_than = null
      }
    } ]
    labels = {
      "type" = "production"
    }
  }
}

variable "docker_cache_repository" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    mode                   = string
    cleanup_policy_dry_run = bool
    cleanup_policies       = list(object({
      id      = string
      action  = string
      condition = object({
        tag_state = string
        older_than = string
      })
    }))
  })
  default = {
    name = "cache"
    description = "Cache repo for kyma-project"
    location = "europe"
    format = "DOCKER"
    immutable_tags = false
    mode = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run = true
    cleanup_policies = [ 
      {
        id = "delete_untagged"
        action = "DELETE"
        condition = {
          tag_state = "UNTAGGED"
          older_than = null
        }
      },
      {
        id = "delete_older_than_week"
        action = "DELETE"
        condition = {
          tag_state = "ANY"
          # Google provider does not support the time units, so we need to provide the time in seconds
          older_than = "604800s" # 604800s = 7 days
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
    cleanup_policy_dry_run = bool
    cleanup_policies       = list(object({
      id      = string
      action  = optional(string)
      condition = optional(object({
        tag_state = optional(string)
        tag_prefixes = optional(list(string), [])
        older_than = optional(string)
      }))
    }))
    labels = map(string)
  })
  default = {
    name = "dev"
    description = "Development images for kyma-projec"
    location = "europe"
    format = "DOCKER"
    immutable_tags = false
    mode = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run = true
    labels = {
      "type" = "development"
    }
    cleanup_policies = [ 
      {
        id = "delete_untagged"
        action = "DELETE"
        condition = {
          tag_state = "UNTAGGED"
          older_than = null
        }
      },
      {
        id = "delete_old_pr"
        action = "DELETE"
        condition = {
          tag_state = "TAGGED"
          
          # Google provider does not support the time units, 
          # so we need to provide the time in seconds.
          older_than = "2592000s" # 2592000s = 720h = 30 days
        }
      }
    ]
  }
}
