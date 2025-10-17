module "github-repository" {
  for_each = var.repositories_rulesets

  source = "../../modules/github-repository"

  repository_name = each.key
  rulesets        = each.value
}