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
