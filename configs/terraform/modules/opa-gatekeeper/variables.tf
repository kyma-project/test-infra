variable "manifests_path" {
  type        = string
  description = "Path to the gatekeeper manifest file to apply to the cluster."
}

variable "constraint_templates_path" {
  type        = list(string)
  description = "Paths to the constraint templates directories."
}

variable "constraints_path" {
  type        = list(string)
  description = "Paths to the constraints directories."
}
# (2025-03-04)