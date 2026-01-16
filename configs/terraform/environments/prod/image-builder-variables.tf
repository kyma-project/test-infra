
variable "dockerhub_credentials" {
  type = object({
    oat_secret_name = string
    username        = string
  })
  default = {
    oat_secret_name = "docker_sap_org_service_auth_token"
    username        = "sapcom"
  }
}

data "google_secret_manager_secret_version" "dockerhub_oat_secret" {
  count   = var.dockerhub_credentials != null ? 1 : 0
  project = var.gcp_project_id
  secret  = var.dockerhub_credentials.oat_secret_name
  version = "latest"
}

# GitHub resources

variable "image_builder_reusable_workflow_ref" {
  type        = string
  description = "The value of GitHub OIDC token job_workflow_ref claim of the image-builder reusable workflow in the test-infra repository. This is used to identify token exchange requests for image-builder reusable workflow."
  default     = "kyma-project/test-infra/.github/workflows/image-builder.yml@refs/heads/main"
}

# GCP resources

variable "image_builder_ado_pat_gcp_secret_manager_secret_name" {
  description = "Name of the secret in GCP Secret Manager that contains the ADO PAT for image-builder to trigger ADO pipeline."
  type        = string
  default     = "image-builder-ado-pat"
}

variable "image_builder_sa_key_restricted_dev_secret_name" {
  description = "Name of the secret in GCP Secret Manager (sap-kyma-prow) that contains the JSON key for the restricted dev image-builder service account."
  type        = string
  default     = "image-builder-sa-key-restricted-dev"
}

variable "image_builder_sa_key_restricted_prod_secret_name" {
  description = "Name of the secret in GCP Secret Manager (sap-kyma-prow) that contains the JSON key for the restricted prod image-builder service account."
  type        = string
  default     = "image-builder-sa-key-restricted-prod"
}

# Variable for image-builder's artifact registries identity
variable "image_builder_kyma-project_identity" {
  description = "Configuration for identity of image-builder in main kyma-project GCP project. It's used to access artifact registries."
  type = object({
    id          = string
    description = string
  })

  default = {
    id          = "azure-pipeline-image-builder"
    description = "OCI image builder running in kyma development service azure pipelines"
  }
}

# Variable for restricted dev image-builder service account
variable "image_builder_kyma-project_identity_restricted_dev" {
  description = "Configuration for restricted dev identity of image-builder in main kyma-project GCP project. It's used to access restricted dev registries."
  type = object({
    id          = string
    description = string
  })

  default = {
    id          = "azure-pipeline-image-builder-restricted-dev"
    description = "OCI image builder service account for restricted development environment in kyma development service azure pipelines"
  }
}

# Variable for restricted prod image-builder service account
variable "image_builder_kyma-project_identity_restricted_prod" {
  description = "Configuration for restricted prod identity of image-builder in main kyma-project GCP project. It's used to access restricted prod registries."
  type = object({
    id          = string
    description = string
  })

  default = {
    id          = "azure-pipeline-image-builder-restricted-prod"
    description = "OCI image builder service account for restricted production environment in kyma development service azure pipelines"
  }
}

variable "dockerhub_mirror" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    mode                       = string
    cleanup_policy_dry_run     = bool
    labels                     = map(string)
    remote_repository_config = optional(object({
      description              = string
      docker_public_repository = string
    }))
    cleanup_policies = optional(list(object({
      id     = string
      action = string
      condition = object({
        tag_state  = optional(string)
        older_than = string
      })
    })), [])
  })
  default = {
    name                       = "dockerhub-mirror"
    description                = "Remote repository mirroring Docker Hub. For more details, see https://github.tools.sap/kyma/oci-image-builder/blob/main/README.md"
    repository_prevent_destroy = false
    location                   = "europe"
    format                     = "DOCKER"
    mode                       = "REMOTE_REPOSITORY"
    cleanup_policy_dry_run     = true
    labels = {
      "type"  = "development"
      "name"  = "dockerhub-mirror"
      "owner" = "neighbors"
    }
    remote_repository_config = {
      description              = "Remote repository mirroring Docker Hub"
      docker_public_repository = "DOCKER_HUB"
    }
    cleanup_policies = [{
      id     = "delete-old-cache"
      action = "DELETE"
      condition = {
        tag_state  = "ANY"
        older_than = "604800s"
      }
    }]
  }
}

variable "docker_cache_repository" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
    location                   = string
    format                     = string
    immutable_tags             = bool
    mode                       = string
    cleanup_policy_dry_run     = bool
    labels                     = map(string)
    cleanup_policies = optional(list(object({
      id     = string
      action = string
      condition = object({
        tag_state  = optional(string)
        older_than = string
      })
    })), [])
  })
  default = {
    name                       = "cache"
    description                = "Cache repo for kyma-project"
    repository_prevent_destroy = false
    location                   = "europe"
    format                     = "DOCKER"
    immutable_tags             = false
    mode                       = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run     = false
    labels = {
      "type"  = "development"
      "name"  = "docker-cache"
      "owner" = "neighbors"
    }
    cleanup_policies = [{
      id     = "delete-old-cache"
      action = "DELETE"
      condition = {
        tag_state  = "ANY"
        older_than = "604800s"
      }
      },
      {
        id     = "delete-untagged"
        action = "DELETE"
        condition = {
          tag_state  = "UNTAGGED"
          older_than = "3600s"
        }
    }]
  }
}
