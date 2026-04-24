variable "telemetry_manager_runner_group_name" {
  type        = string
  default     = "telemetry-manager-runners"
  description = "Name of the GitHub Actions runner group for telemetry-manager"
}

variable "telemetry_manager_hosted_runner" {
  type = object({
    name         = string
    size         = string
    max_runners  = number
    image_id     = string
    image_source = string
  })

  default = {
    name         = "telemetry-4core"
    size         = "4-core"
    max_runners  = 30
    image_id     = "2306"
    image_source = "github"
  }

  description = "Configuration for the telemetry-manager GitHub-hosted larger runner"
}
