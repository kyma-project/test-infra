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

# Variables for Docker Hub Mirror configuration
variable "dockerhub_mirror_repository_id" {
  description = "Name of the Docker Hub mirror repository"
  type        = string
  default     = "dockerhub-mirror"
}

variable "dockerhub_mirror_description" {
  description = "Description of the Docker Hub mirror repository"
  type        = string
  default     = "Remote repository mirroring Docker Hub. For more details, see https://github.tools.sap/kyma/oci-image-builder/blob/main/README.md"
}

variable "dockerhub_mirror_location" {
  description = "Location of the Docker Hub mirror repository"
  type        = string
  default     = "europe"
}

variable "dockerhub_mirror_member" {
  description = "IAM member to assign the role to (service account)"
  type        = string
  default     = "serviceAccount:azure-pipeline-image-builder@kyma-project.iam.gserviceaccount.com"
}

# Variable for the Docker Hub mirror cleanup age
variable "dockerhub_mirror_cleanup_age" {
  description = "Age after which to clean up images in the Docker Hub mirror repository"
  type        = string
  default     = "730d"  # 730 days = 2 years
}
