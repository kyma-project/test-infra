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
    project_id = "sap-kyma-prow"
  }

  description = "Terraform executor gcp service account details."
}

variable "gcp_region" {
  type    = string
  default = "europe-west4"
}

variable "gcp_project_id" {
  type    = string
  default = "sap-kyma-prow"
}

variable "workloads_project_id" {
  type        = string
  default     = "sap-kyma-prow-workloads"
  description = "Additional Google Cloud project ID to grant the IAM permissions to terraform-executor service account."
}

variable "tekton_k8s_config_path" {
  type        = string
  default     = "/kubeconfigs/tekton-config"
  description = "Path to kubeconfig file ot use to connect to tekton k8s cluster."
}

variable "tekton_k8s_config_context" {
  type        = string
  default     = "gke_sap-kyma-prow_europe-west4_tekton"
  description = "Context to use to connect to tekton k8s cluster."
}

variable "trusted_workloads_k8s_config_path" {
  type        = string
  default     = "/kubeconfigs/trusted-workloads-config"
  description = "Path to kubeconfig file ot use to connect to trusted workloads k8s cluster."
}

variable "trusted_workloads_k8s_config_context" {
  type        = string
  default     = "gke_sap-kyma-prow_europe-west3_trusted-workload-kyma-prow"
  description = "Context to use to connect to trusted workloads k8s cluster."
}

variable "untrusted_workloads_k8s_config_path" {
  type        = string
  default     = "/kubeconfigs/untrusted-workloads-config"
  description = "Path to kubeconfig file ot use to connect to untrusted workloads k8s cluster."
}

variable "untrusted_workloads_k8s_config_context" {
  type        = string
  default     = "gke_sap-kyma-prow_europe-west3_untrusted-workload-kyma-prow"
  description = "Context to use to connect to untrusted workloads k8s cluster."
}
