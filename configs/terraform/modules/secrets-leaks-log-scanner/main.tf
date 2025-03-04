terraform {
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

variable "prow_pubsub_topic_name" {
  type    = string
  default = "prowjobs"
}

# Used to retrieve project_number later
data "google_project" "project" {
  provider = google
}

data "google_storage_bucket" "kyma_prow_logs" {
  name = "kyma-prow-logs"
}

data "google_secret_manager_secret" "gh_tools_kyma_bot_token" {
  secret_id = "trusted_default_kyma-bot-github-sap-token"
}

variable "slack_message_sender_url" {
  type = string
}
# (2025-03-04)