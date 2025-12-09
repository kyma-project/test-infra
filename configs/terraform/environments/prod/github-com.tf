# TODO:
# These actions variables must exist in github in order to let terraform executor run.
# They are required for getting terraform executor github personal access token from gcp secret manager.
# We must develop a solution for creating a minimal initial setup to let terraform executor apply our whole config.

# GCP project id that contains secret manager with secrets
resource "github_actions_organization_variable" "gcp_kyma_project_project_id" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "GCP_KYMA_PROJECT_PROJECT_ID"
  value         = var.gcp_project_id
}

data "github_organization" "kyma-project" {
  provider = github.kyma_project
  name     = "kyma-project"
}

variable "kyma-bot-gcp-secret-name-public-github-token" {
  type        = string
  default     = "kyma-bot-github-public-repo-token"
  description = "GCP Secret Manager secret name for public GitHub token used by kyma bot"
}

import {
  id = "projects/${var.gcp_project_id}/secrets/${var.kyma-bot-gcp-secret-name-public-github-token}"
  to = google_secret_manager_secret.kyma-bot-public-github-token
}

# kyma-bot-public-github-token is not connected to one particular application.
# It is used by multiple infrastructure components and applications and therefore it is created here as a variable realted to github.com instance.
resource "google_secret_manager_secret" "kyma-bot-public-github-token" {
  project   = var.gcp_project_id
  secret_id = var.kyma-bot-gcp-secret-name-public-github-token
  replication {
    auto {}
  }
}