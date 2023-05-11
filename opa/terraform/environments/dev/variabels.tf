variable "tekton_gatekeeper_manifest_path" {
  type = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
}

variable "trusted_workloads_gatekeeper_manifest_path" {
  type = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
}

variable "untrusted_workloads_gatekeeper_manifest_path" {
  type = string
  default = "../../../gatekeeper/deployments/gatekeeper.yaml"
}
