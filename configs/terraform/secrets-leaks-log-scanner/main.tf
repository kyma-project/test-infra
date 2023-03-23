terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-project"
    prefix = "secret-leaks-log-scanner"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.58.0"
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

provider "google" {
  project = var.gcp_project_id
  region  = "europe-west3"
  zone    = "europe-west3-a"
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

data "google_secret_manager_secret" "common_slack_bot_token" {
  secret_id = "common-slack-bot-token"
}
