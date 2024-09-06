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