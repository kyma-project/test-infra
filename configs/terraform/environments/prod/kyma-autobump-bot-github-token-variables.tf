variable "kyma_autobump_bot_github_token_secret_name" {
  type        = string
  description = "Name of the kyma-autobump-bot-github-token secret. This secret is used by automatic bumpers to interact with GitHub."
  default     = "kyma-autobump-bot-github-token"
}
