variable "image_builder_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })

  default = {
    name      = "image-builder"
    namespace = "default"
  }
  description = <<-EOT
    Details of image builder k8s service account.
    name: Name of the image builder k8s service account.
    namespace: Namespace of the image builder k8s service account.
    EOT
}

variable "external-secrets-sa-trusted-cluster" {
  type = object({
    name      = string
    namespace = string
  })
  default = {
    name      = "secret-manager-trusted"
    namespace = "external-secrets"
  }
  description = <<-EOT
    Details of external secrets service account.
    name: Name of the external secret controller service account.
    namespace: Namespace of the external secret controller service account.
    EOT
}

variable "external-secrets-sa-untrusted-cluster" {
  type = object({
    name      = string
    namespace = string
  })
  default = {
    name      = "secret-manager-untrusted"
    namespace = "external-secrets"
  }
  description = <<-EOT
    Details of external secrets service account.
    name: Name of the external secret controller service account.
    namespace: Namespace of the external secret controller service account.
    EOT
}

variable "signify-dev-secret-name" {
  type        = string
  description = "Name of the signify dev secret."
  default     = "signify-dev-secret"
}

variable "signify-prod-secret-name" {
  type        = string
  description = "Name of the signify dev secret."
  default     = "signify-prod-secret"
}
