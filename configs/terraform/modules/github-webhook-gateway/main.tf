terraform {
  required_version = ">= 1.6.1"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.76.0"
    }
  }
}

variable "gcp_project_id" {
  type    = string
  default = "sap-kyma-prow"
}

variable "pubsub_topic_name" {
  type    = string
  default = "issue-labeled"
}

data "google_secret_manager_secret" "gh_tools_kyma_bot_token" {
  secret_id = "trusted_default_kyma-bot-github-sap-token"
}

data "google_secret_manager_secret" "webhook_token" {
  secret_id = "sap-tools-github-backlog-webhook-secret"
}
