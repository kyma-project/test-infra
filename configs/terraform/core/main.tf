# Creation of this secret must be part of bootstrap process of kyma infrastrucutre. Creation must be done using human user with admin rights.
#resource "github_actions_secret" "kyma-bot-github-token" {
#  provider = "github.kyma-project"
#  repository       = "test-infra"
#  secret_name      = "GITHUB_TOKEN"
#  plaintext_value  = var.github_token
#}
