
variable "kyma_bot_github_sap_token_secret_name" {
  type        = string
  description = "Name of the kyma-autobump-bot-github-token secret in the Google's Secret Manager. This secret is used by automatic bumpers to interact with GitHub."
  default     = "kyma-autobump-bot-github-token"
}

