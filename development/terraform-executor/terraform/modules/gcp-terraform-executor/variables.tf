variable "terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })
  description = "Terraform executor k8s service account details."
}

variable "terraform_executor_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })
  description = "Terraform executor gcp service account details."
}
