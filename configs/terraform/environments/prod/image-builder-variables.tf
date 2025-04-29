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
    labels                 = map(string)
    cleanup_policies       = optional(list(object({
      id        = string
      action    = string
      condition = object({
        tag_state    = optional(string)
        older_than   = string
      })
    })), [])
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
    labels = {
      "type"  = "development"
      "name"  = "docker-cache"
      "owner" = "neighbors"
    }
    cleanup_policies = [{
      id        = "delete-old-cache"
      action    = "DELETE"
      condition = {
        tag_state  = "ANY"
        older_than = "604800s"
      }
    }]
  }
}

variable "dockerhub_mirror" {
  type = object({
    name                   = string
    description            = string
    location               = string
    format                 = string
    immutable_tags         = bool
    mode                   = string
    cleanup_policy_dry_run = bool
    labels                 = map(string)
    remote_repository_config = optional(object({
      description = string
      docker_repository = object({
        public_repository = string
      })
    }))
    cleanup_policies       = optional(list(object({
      id        = string
      action    = string
      condition = object({
        tag_state    = optional(string)
        older_than   = string
      })
    })), [])
  })
  default = {
    name                   = "dockerhub-mirror"
    description            = "Remote repository mirroring Docker Hub. For more details, see https://github.tools.sap/kyma/oci-image-builder/blob/main/README.md"
    location               = "europe"
    format                 = "DOCKER"
    immutable_tags         = false
    mode                   = "REMOTE_REPOSITORY"
    cleanup_policy_dry_run = true
    labels = {
      "type"  = "development"
      "name"  = "dockerhub-mirror"
      "owner" = "neighbors"
    }
    remote_repository_config = {
      description = "Remote repository mirroring Docker Hub"
      docker_repository = {
        public_repository = "DOCKER_HUB"
      }
    }
    cleanup_policies = [{
      id        = "delete-old-cache"
      action    = "DELETE"
      condition = {
        tag_state  = "ANY"
        older_than = "604800s"
      }
    }]
  }
}
