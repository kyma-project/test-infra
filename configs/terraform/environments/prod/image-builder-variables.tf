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

# GitHub resources

variable "image_builder_reusable_workflow_name" {
  type        = string
  description = "Name of the image-builder reusable workflow in the test-infra repository."
  default = "image-builder"
}

variable "image_builder_reusable_workflow_ref" {
  type        = string
  description = "Name of the image-builder reusable workflow in the test-infra repository."
  default     = "kyma-project/test-infra/.github/workflows/image-builder.yml@refs/heads/main"
}

# GCP resources

variable "image_builder_ado_pat_gcp_secret_manager_secret_name" {
  description = "Name of the secret in GCP Secret Manager that contains the ADO PAT for image-builder to trigger ADO pipeline."
  type        = string
  default     = "image-builder-ado-pat"
}

variable "image_builder_gh_workflow_service_account" {
  description = "Service account used by image-builder reusable workflow to access GCP secret manager. Reusable workflow is defined in the test-infra repository."

  type = object({
    id         = string
    project_id = string
  })

  default = {
    id         = "image-builder-gh-workflow"
    project_id = "sap-kyma-prow"
  }
}