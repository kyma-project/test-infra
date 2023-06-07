variable "pjtester_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })

  default = {
    name      = "pjtester"
    namespace = "default"
  }
  description = <<-EOT
    Details of image builder k8s service account.
    name: Name of the image builder k8s service account.
    namespace: Namespace of the image builder k8s service account.
    EOT
}

variable "pjtester-kubeconfig-secret-name" {
  type        = string
  description = "Name of the pjtester secret."
  default     = "pjtester-kubeconfig"
}

variable "pjtester-github-token-secret-name" {
  type        = string
  description = "Name of the pjtester GitHub token secret."
  default     = "pjtester-github-oauth-token"
}
