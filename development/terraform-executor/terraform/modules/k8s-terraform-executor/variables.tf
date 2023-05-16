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

variable "k8s_config_path" {
  type        = string
  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
}

variable "k8s_config_context" {
  type        = string
  description = "Context to use to connect to managed k8s cluster."
}
