variable "kyma_bot_github_token_secret_name" {
  type        = string
  description = "Name of the kyma-bot-github-token secret. This secret is used by automation to interact with GitHub."
  default     = "kyma-bot-github-token"
}
variable "kyma_autobump_bot_github_token_secret_name" {
  type        = string
  description = "Name of the kyma-autobump-bot-github-token secret. This secret is used by automatic bumpers to interact with GitHub."
  default     = "kyma-autobump-bot-github-token"
}

# TODO(kacpermalachowski): Rename to kyma_autobump_bot_github_token_secret_name after Prow removal
variable "kyma_autobump_bot_github_token_sm_secret_name" {
  type        = string
  description = "Name of the kyma-autobump-bot-github-token secret in the Google's Secret Manager. This secret is used by automatic bumpers to interact with GitHub."
  default     = "workloads_default_kyma-autobump-bot-github-token"
}

variable "kyma_bot_github_sap_token_secret_name" {
  type        = string
  description = "Name of the kyma-bot-github-sap-token secret. This is used by automation to interact with SAP GitHub instance."
  default     = "kyma-bot-github-sap-token"
}

variable "kyma_guard_bot_github_token_secret_name" {
  type        = string
  description = "Name of the kyma-guard-bot-github-token secret. This secret is used by automation to get GitHub contexts statuses."
  default     = "kyma-guard-bot-github-token"
}
