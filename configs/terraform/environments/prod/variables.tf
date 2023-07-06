variable "gcp_region" {
  type        = string
  default     = "europe-west4"
  description = "Default Google Cloud region to create resources."
}

variable "gcp_project_id" {
  type        = string
  default     = "sap-kyma-prow"
  description = "Google Cloud project to create resources."
}

variable "workloads_project_id" {
  type        = string
  default     = "sap-kyma-prow-workloads"
  description = "Additional Google Cloud project ID."
}

variable "gatekeeper_manifest_path" {
  type        = string
  default     = "../../../../opa/gatekeeper/deployments/gatekeeper.yaml"
  description = "Path to the Gatekeeper yaml manifest file. This file will be applied to the k8s cluster to install gatekeeper."
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
    location = "europe-west3"
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

variable "external_secrets_k8s_sa_trusted_cluster" {
  type = object({
    name      = string
    namespace = string
  })
  default = {
    name      = "secret-manager-trusted"
    namespace = "external-secrets"
  }
  description = <<-EOT
    Details of external secrets service account. This is service account used as identity for external secrets controller.
    name: Name of the external secret controller service account.
    namespace: Namespace of the external secret controller service account.
    EOT
}

variable "external_secrets_k8s_sa_untrusted_cluster" {
  type = object({
    name      = string
    namespace = string
  })
  default = {
    name      = "secret-manager-untrusted"
    namespace = "external-secrets"
  }
  description = <<-EOT
    Details of external secrets service account. This is service account used as identity for external secrets controller.
    name: Name of the external secret controller service account.
    namespace: Namespace of the external secret controller service account.
    EOT
}
