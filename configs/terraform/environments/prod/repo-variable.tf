resource "github_actions_variable" "terraform_runner_github_app_secret_name" {
  provider      = github
  repository    = "test-infra"
  variable_name = "testing-variable"
  value         = "test-value"
}