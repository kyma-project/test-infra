variable "dev_modules_internal_repository" {
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
    name                   = "dev-modules-internal"
    description            = "Kyma modules created by submission pipeline on pull requests."
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    mode                   = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run = true
    # Google provider does not support the time units,
    # so we need to provide the time in seconds.
    # Time after which the images will be deleted.
    labels = {
      "type"  = "development"
      "name"  = "dev-modules-internal"
      "owner" = "neighbors"
    }
  }
}# (2025-03-04)