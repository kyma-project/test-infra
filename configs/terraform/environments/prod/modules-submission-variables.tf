variable "dev_modules_internal_repository" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    mode                   = string
    cleanup_policy_dry_run = bool
    labels                 = map(string)
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
}

variable "dev_kyma_modules_repository" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
  })
  default = {
    name                       = "dev-kyma-modules"
    description                = "Development Kyma modules"
    repository_prevent_destroy = false
  }
}

variable "kyma_modules_repository" {
  type = object({
    name                       = string
    description                = string
    type                       = string
    reader_serviceaccounts     = list(string)
    repository_prevent_destroy = bool
  })
  default = {
    name        = "kyma-modules"
    description = "Production Kyma modules"
    type        = "production"
    reader_serviceaccounts = [
      "klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com",
      "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com",
      "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"
    ]
    repository_prevent_destroy = true
  }
}