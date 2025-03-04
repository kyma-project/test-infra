variable "pjtester_kubeconfig_secret_name" {
  type        = string
  description = "Name of the pjtester secret. This secret contains the kubeconfig for the prow cluster. Pjtester will use it to schedule test prowjob."
  default     = "pjtester-kubeconfig"
}

variable "pjtester_github_token_secret_name" {
  type        = string
  description = "Name of the pjtester GitHub token secret. This secret will be used to create a GitHub status for the test prowjob."
  default     = "pjtester-github-oauth-token"
}
# (2025-03-04)