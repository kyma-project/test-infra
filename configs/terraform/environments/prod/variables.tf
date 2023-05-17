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

variable "tekton_gatekeeper_manifest_path" {
  type    = string
  default = "../../../../opa/gatekeeper/deployments/gatekeeper.yaml"
}

variable "tekton_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  default = {
    name     = "tekton"
    location = "europe-west4"
  }

  description = "Name of the managed k8s cluster to apply the manifest to."
}

variable "trusted_workload_gatekeeper_manifest_path" {
  type    = string
  default = "../../../../opa/gatekeeper/deployments/gatekeeper.yaml"
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

  description = "Name of the managed k8s cluster to apply the manifest to."
}

variable "untrusted_workload_gatekeeper_manifest_path" {
  type    = string
  default = "../../../../opa/gatekeeper/deployments/gatekeeper.yaml"
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

  description = "Name of the managed k8s cluster to apply the manifest to."
}

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
