variable "pjtester_kubeconfig_secret_name" {
  type        = string
  description = "Name of the pjtester secret."
  default     = "pjtester-kubeconfig"
}

variable "pjtester_github_token_secret_name" {
  type        = string
  description = "Name of the pjtester GitHub token secret."
  default     = "pjtester-github-oauth-token"
}
