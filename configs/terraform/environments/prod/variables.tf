variable "gcp_region" {
  type        = string
  default     = "europe-west4"
  description = "Default Google Cloud region to create resources."
}

variable "gcp_scheduler_region" {
  type        = string
  default     = "europe-west3"
  description = "Additional Google Cloud Region to create resources."
}

variable "gcp_project_id" {
  type        = string
  default     = "sap-kyma-prow"
  description = "Google Cloud project to create resources."
}

variable "prow_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  default = {
    name     = "prow"
    location = "europe-west3-a"
  }

  description = "Details of the prow k8s cluster."
}

variable "trusted_workload_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  default = {
    name     = "trusted-workload-kyma-prow"
    location = "europe-west4"
  }

  description = "Details of the trusted-workload k8s cluster."
}

variable "untrusted_workload_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  default = {
    name     = "untrusted-workload-kyma-prow"
    location = "europe-west3"
  }

  description = "Details of the untrusted-workload k8s cluster."
}

variable "kyma_project_gcp_region" {
  type        = string
  description = "Default Google Cloud region to create resources in kyma-project"
  default     = "europe-west4"
}

variable "kyma_project_gcp_project_id" {
  type        = string
  description = "Google Cloud project to create resources"
  default     = "kyma-project"
}

variable "automated_approver_deployment_path" {
  type        = string
  description = "Path to the automated-approver deployment file"
  default     = "../../../../prow/cluster/components/automated-approver_external-plugin.yaml"
}

variable "automated_approver_rules_path" {
  type        = string
  description = "Path to the automated-approver rules file"
  default     = "../../../../configs/automated-approver-rules.yaml"
}


variable "kyma-project-github-org" {
  type        = string
  default     = "kyma-project"
  description = "The GitHub organization where the Kyma project is hosted"
}
