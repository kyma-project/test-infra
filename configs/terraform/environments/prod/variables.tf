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

variable "tekton_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  default = {
    name     = "tekton"
    location = "europe-west4"
  }

  description = "Details of the tekton k8s cluster."
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

variable "terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })

  default = {
    name      = "terraform-executor"
    namespace = "default"
  }

  description = "Details of terraform executor k8s service account."
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

  description = "Details of terraform executor gcp service account."
}

variable "constraint_templates_path" {
  type    = string
  default = "../../../../opa/gatekeeper/constraint-templates/**.yaml"

  description = "Path to the constraint templates directory."
}

variable "tekton_constraints_path" {
  type    = string
  default = "../../../../tekton/deployments/gatekeeper-constraints/**.yaml"

  description = "Path to the tekton cluster constraints directory."
}

variable "prow_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/prow/**.yaml"

  description = "Path to the trusted workloads cluster constraints directory."
}

variable "trusted_workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/trusted/**.yaml"

  description = "Path to the trusted workloads cluster constraints directory."
}

variable "untrusted_workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/untrusted/**.yaml"

  description = "Path to the untrusted workloads cluster constraints directory."
}

variable "workloads_constraints_path" {
  type    = string
  default = "../../../../prow/cluster/resources/gatekeeper-constraints/workloads/**.yaml"

  description = "Path to both workload clusters constraints directory."
}
