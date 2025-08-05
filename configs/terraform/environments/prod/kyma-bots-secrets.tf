 resource "github_actions_variable" "kyma_autobump_bot_github_token_secret_name" {
  provider      = github.kyma_project
  repository    = data.github_repository.test_infra.name
  variable_name = "KYMA_AUTOBUMP_BOT_GITHUB_SECRET_NAME"
  value         = var.kyma_bot_github_sap_token_secret_name
}
