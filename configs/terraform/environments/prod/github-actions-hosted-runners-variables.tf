variable "telemetry_manager_runner_group_name" {
  type        = string
  default     = "telemetry-manager-runners"
  description = "Name of the GitHub Actions runner group for telemetry-manager"
}

variable "telemetry_manager_hosted_runner_name" {
  type        = string
  default     = "telemetry-4core"
  description = "Name and label of the GitHub-hosted larger runner for telemetry-manager"
}

variable "telemetry_manager_hosted_runner_size" {
  type        = string
  default     = "4-core"
  description = "Machine size for the hosted runner (e.g., 4-core, 8-core, 16-core)"
}

variable "telemetry_manager_hosted_runner_max_runners" {
  type        = number
  default     = 30
  description = "Maximum number of runners that can be scaled up simultaneously"
}

variable "telemetry_manager_hosted_runner_image_id" {
  type        = string
  default     = "2306"
  description = "Image ID for the hosted runner (2306 = Ubuntu Latest 24.04)"
}

variable "telemetry_manager_hosted_runner_image_source" {
  type        = string
  default     = "github"
  description = "Image source for the hosted runner (github, partner, or custom)"
}
