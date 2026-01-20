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
          tag_state    = "TAGGED"
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
      docker_public_repository = optional(string)
      docker_custom_repository = optional(string)
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
      docker_custom_repository = "https://cgr.dev"
    }
  }
}

variable "restricted_prod" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    mode                       = string
    type                       = string
    cleanup_policy_dry_run     = bool
    virtual_repository_config = object({
      upstream_policies = optional(list(object({
        id         = string
        repository = string
        priority   = number
      })))
    })
  })
  default = {
    name                       = "restricted-prod"
    description                = "Virtual repository for restricted production images"
    repository_prevent_destroy = true
    location                   = "europe"
    format                     = "DOCKER"
    mode                       = "VIRTUAL_REPOSITORY"
    type                       = "production"
    cleanup_policy_dry_run     = false
    virtual_repository_config = {
      upstream_policies = [
        {
          id         = "kyma-restricted-images-prod"
          repository = "projects/kyma-project/locations/europe/repositories/kyma-restricted-images-prod"
          priority   = 100
        },
        {
          id         = "chainguard-cache"
          repository = "projects/kyma-project/locations/europe/repositories/chainguard-cache"
          priority   = 50
        }
      ]
    }
  }
}

variable "restricted_dev" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    mode                       = string
    type                       = string
    cleanup_policy_dry_run     = bool
    virtual_repository_config = object({
      upstream_policies = optional(list(object({
        id         = string
        repository = string
        priority   = number
      })))
    })
  })
  default = {
    name                       = "restricted-dev"
    description                = "Virtual repository for restricted development images"
    repository_prevent_destroy = true
    location                   = "europe"
    format                     = "DOCKER"
    mode                       = "VIRTUAL_REPOSITORY"
    type                       = "development"
    cleanup_policy_dry_run     = false
    virtual_repository_config = {
      upstream_policies = [
        {
          id         = "kyma-restricted-images-dev"
          repository = "projects/kyma-project/locations/europe/repositories/kyma-restricted-images-dev"
          priority   = 100
        },
        {
          id         = "chainguard-cache"
          repository = "projects/kyma-project/locations/europe/repositories/chainguard-cache"
          priority   = 50
        }
      ]
    }
  }
}

variable "chainguard_pull_token_secret_name" {
  type        = string
  description = "Name of the Secret Manager secret containing Chainguard pull token password"
  default     = "chainguard_auth_token"
}

variable "restricted_registry_iam_groups" {
  type = object({
    prod_read             = string
    prod_read_group_name  = string
    prod_write            = string
    prod_write_group_name = string
    dev_read              = string
    dev_read_group_name   = string
    dev_write             = string
    dev_write_group_name  = string
  })
  default = {
    prod_read             = "kyma-restricted-registry-prod-read@sap.com"
    prod_read_group_name  = "groups/00xvir7l1dtv8ew"
    prod_write            = "kyma-restricted-registry-prod-write@sap.com"
    prod_write_group_name = "groups/01egqt2p1a2johw"
    dev_read              = "kyma-restricted-registry-dev-read@sap.com"
    dev_read_group_name   = "groups/0184mhaj2tdduaw"
    dev_write             = "kyma-restricted-registry-dev-write@sap.com"
    dev_write_group_name  = "groups/023ckvvd0rmgw6a"
  }
  description = "Google Cloud Identity groups for Restricted Registry access control"
}

variable "restricted_registry_hierarchical_groups" {
  type = object({
    security_scanners            = string
    security_scanners_group_name = string
    developers                   = string
    developers_group_name        = string
    markets_delivery             = string
    markets_delivery_group_name  = string
    image_builder                = string
    image_builder_group_name     = string
    image_signer                 = string
    image_signer_group_name      = string
  })
  default = {
    security_scanners            = "kyma-restricted-registry-security-scanners@sap.com"
    security_scanners_group_name = "groups/00meukdy10z0qky"
    developers                   = "kyma-restricted-registry-developers@sap.com"
    developers_group_name        = "groups/02koq65619rujz4"
    markets_delivery             = "kyma-restricted-registry-markets-delivery@sap.com"
    markets_delivery_group_name  = "groups/03fwokq04jj0scm"
    image_builder                = "kyma-restricted-registry-image-builder@sap.com"
    image_builder_group_name     = "groups/01fob9te0l19xwg"
    image_signer                 = "kyma-restricted-registry-image-signer@sap.com"
    image_signer_group_name      = "groups/01yyy98l2u8q924"
  }
  description = "Hierarchical Cloud Identity groups for organizing service accounts with Restricted Registry access"
}
