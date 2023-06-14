variable "terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })
  description = "Details of terraform executor k8s service account."
}

variable "terraform_executor_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })
  description = "Details of terraform executor gcp service account."
}

variable "external_secrets_sa" {
  type = object({
    name      = string
    namespace = string
  })
  description = <<-EOT
    Details of external secrets service account.
    name: Name of the external secret controller service account.
    namespace: Namespace of the external secret controller service account.
    EOT
}
