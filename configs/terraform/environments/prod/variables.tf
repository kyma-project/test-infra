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

variable "secrets_rotator_name" {
  type        = string
  description = "Name of the secrets rotator application."
  default     = "secrets-rotator"
}


variable "secrets_rotator_cloud_run_listen_port" {
  type        = number
  description = "Port on which the secrets rotator services listens."
  default     = 8080
}

variable "service_account_keys_rotator_service_name" {
  type        = string
  description = "Name of the service account keys rotator service instance."
}

variable "service_account_keys_rotator_account_id" {
  type        = string
  default     = "sa-keys-rotator"
  description = "Service account ID of the service account keys rotator."
}

variable "service_account_keys_rotator_image" {
  type        = string
  description = "Image of the service account keys rotator."
  validation {
    condition     = can(regex("^europe-docker.pkg.dev/kyma.*", var.service_account_keys_rotator_image))
    error_message = "The service account keys rotator image must be hosted in the Kyma Google Artifact Registry."
  }
}

variable "secret_manager_notifications_topic" {
  type        = string
  description = "Name of the topic to which the service account keys rotator subscribes to."
  default     = "secret-manager-notifications"
}

variable "service_account_keys_cleaner_service_name" {
  type        = string
  description = "Name of the service account keys cleaner service instance."
}

variable "service_account_keys_cleaner_account_id" {
  type        = string
  default     = "sa-keys-cleaner"
  description = "Service account ID of the service account keys cleaner."
}

variable "service_account_keys_cleaner_image" {
  type        = string
  description = "Image of the service account keys cleaner."
  validation {
    condition     = can(regex("^europe-docker.pkg.dev/kyma.*", var.service_account_keys_cleaner_image))
    error_message = "The service account keys cleaner image must be hosted in the Kyma Google Artifact Registry."
  }
}

variable "service_account_key_latest_version_min_age" {
  type        = number
  description = "Minimum age in hours the service account key latest version exist, before old version to be deleted."
}

variable "service_account_keys_cleaner_scheduler_cron_schedule" {
  type        = string
  description = "Cron schedule for the service account keys cleaner scheduler."
}

variable "prow_cluster_ip_range" {
  type        = string
  default     = "10.8.0.0/14"
  description = "Internal IP address range for pods."
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
