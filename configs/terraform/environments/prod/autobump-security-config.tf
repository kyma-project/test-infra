# ==============================================================================
# Kyma Autobump Bot GitHub Token Secret Access
# ==============================================================================
# This configuration grants the autobump-security-config workflow access to
# the kyma-bot-github-public-repo-token secret in GCP Secret Manager.
# ==============================================================================

# Grant autobump-security-config workflow access to read the kyma-bot public GitHub token secret
resource "google_secret_manager_secret_iam_member" "autobump_security_config_workflow_token_reader" {
  project   = var.gcp_project_id
  secret_id = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "principalSet://iam.googleapis.com/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/attribute.workflow_ref/${data.github_organization.kyma_project.name}/${data.github_repository.test_infra.name}/.github/workflows/autobump-security-config.yaml@refs/heads/main"
}

resource "github_actions_variable" "kyma_autobump_bot_github_token_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = "KYMA_AUTOBUMP_BOT_GITHUB_SECRET_NAME"
  value         = google_secret_manager_secret.kyma_bot_public_github_token.secret_id
}
