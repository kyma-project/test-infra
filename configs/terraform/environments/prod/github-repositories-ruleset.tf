module "github-repository" {
  for_each = toset(var.repository_names)

  source = "../../modules/github-repository"

  repository_name = each.value
  rulesets        = var.rulesets
}