variable "terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })

  default = {
    name      = "terraform-executor"
    namespace = "default"
  }

  description = "Terraform executor k8s service account details."
}

variable "terraform_executor_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })

  default = {
    id         = "terraform-executor"
    project_id = "sap-kyma-neighbors-dev"
  }

  description = "Terraform executor gcp service account details."
}

variable "gcp_region" {
  type    = string
  default = "europe-west4"
}

variable "gcp_project_id" {
  type    = string
  default = "sap-kyma-neighbors-dev"
}

variable "k8s_config_path" {
  type        = string
  default     = "~/.kube/config"
  description = "Path to kubeconfig file ot use to connect to managed k8s cluster."
}

variable "k8s_config_context" {
  type        = string
  description = "Context to use to connect to managed k8s cluster."
}
