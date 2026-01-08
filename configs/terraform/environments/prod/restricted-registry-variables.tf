variable "kyma_restricted_images_prod" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    immutable_tags             = bool
    mode                       = string
    type                       = string
    cleanup_policy_dry_run     = bool
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
    name                       = "kyma-restricted-images-prod"
    description                = "Production restricted images for kyma-project"
    repository_prevent_destroy = true
    location                   = "europe"
    format                     = "DOCKER"
    immutable_tags             = false
    mode                       = "STANDARD_REPOSITORY"
    type                       = "production"
    cleanup_policy_dry_run     = false
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

variable "kyma_restricted_images_dev" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    immutable_tags             = bool
    mode                       = string
    type                       = string
    cleanup_policy_dry_run     = bool
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
    name                       = "kyma-restricted-images-dev"
    description                = "Development restricted images for kyma-project"
    repository_prevent_destroy = true
    location                   = "europe"
    format                     = "DOCKER"
    immutable_tags             = false
    mode                       = "STANDARD_REPOSITORY"
    type                       = "development"
    cleanup_policy_dry_run     = false
    cleanup_policies = [
      {
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
          older_than   = "2592000s"
          tag_prefixes = ["PR-"]
        }
      }
    ]
  }
}

variable "chainguard_cache" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    mode                       = string
    type                       = string
    cleanup_policy_dry_run     = bool
    remote_repository_config = object({
      description              = string
      docker_public_repository = string
      upstream_username        = optional(string)
      upstream_password_secret = optional(string)
    })
  })
  default = {
    name                       = "chainguard-cache"
    description                = "Remote repository for Chainguard pull-through cache"
    repository_prevent_destroy = true
    location                   = "europe"
    format                     = "DOCKER"
    mode                       = "REMOTE_REPOSITORY"
    type                       = "production"
    cleanup_policy_dry_run     = false
    remote_repository_config = {
      description              = "Chainguard upstream repository"
      docker_public_repository = "CHAINGUARD"
    }
  }
}
