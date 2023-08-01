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

data "google_secret_manager_secret" "common_slack_bot_token" {
  secret_id = "common-slack-bot-token"
}


variable "release_cluster_created_pubsub_topic_name" {
  type    = string
  default = "release-cluster-created"
}


variable "release_cluster_created_could_run_path" {
  type    = string
  default = "/release-cluster-created"
}
