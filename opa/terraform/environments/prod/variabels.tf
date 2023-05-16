variable "tekton_gatekeeper_manifest_path" {
  type    = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
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

variable "trusted_workloads_gatekeeper_manifest_path" {
  type    = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
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

variable "untrusted_workloads_gatekeeper_manifest_path" {
  type    = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
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
