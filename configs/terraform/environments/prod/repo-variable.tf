resource "github_actions_variable" "terraform_runner_github_app_secret_name" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "testing_variable"
  value         = "test-value"
}