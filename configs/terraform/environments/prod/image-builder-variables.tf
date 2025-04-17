variable "signify_dev_secret_name" {
  type        = string
  description = "Name of the signify dev secret. This secret is used by image-builder to sign OCI images with signify development service."
  default     = "signify-dev-secret"
}

variable "signify_prod_secret_name" {
  type        = string
  description = "Name of the signify dev secret. This secret is used by image-builder to sign OCI images with signify production service."
  default     = "signify-prod-secret"
}

variable "dockerhub_oat_secret_name" {
  type        = string
  description = "Name of the Docker Hub OAT secret. This secret is used by image-builder's dockerhub_mirror to access Docker Hub."
  default = "docker_sap_org_service_auth_token"
}

variable "dockerhub_username" {
  type        = string
  description = "Name of the Docker Hub username. This is used by image-builder's dockerhub_mirror to access Docker Hub."
  default = "sapcom"
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

# Variable for Docker Hub Mirror configuration
variable "dockerhub_mirror" {
  description = "Configuration for the Docker Hub mirror repository"
  type = object({
    repository_id = string
    description   = string
    location      = string
    cleanup_age   = string
    mode          = string
    mode          = string
  })

  default = {
    repository_id = "dockerhub-mirror"
    description   = "Remote repository mirroring Docker Hub. For more details, see https://github.tools.sap/kyma/oci-image-builder/blob/main/README.md"
    location      = "europe"
    cleanup_age   = "63072000s" # 63072000s = 730 days = 2 years
    mode          = "REMOTE_REPOSITORY"
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
    cache_images_max_age   = string
  })
  default = {
    name                   = "cache"
    description            = "Cache repo for kyma-project"
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    mode                   = "STANDARD_REPOSITORY"
    cleanup_policy_dry_run = false
    # Google provider does not support the time units,
    # so we need to provide the time in seconds.
    # Time after which the images will be deleted.
    cache_images_max_age = "604800s" # 604800s = 7 days
  }
}

variable "remote_repository_config" {
  type = object({
    description = optional(string)
    docker_repository = optional(object({
      public_repository = string
    }))
    upstream_credentials = optional(object({
      username_password_credentials = optional(object({
        username                = string
        password_secret_version = string
      }))
    }))
  })
  default = null
}
