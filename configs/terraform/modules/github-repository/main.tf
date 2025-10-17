module "ruleset" {
  for_each = { for idx, ruleset in var.rulesets : idx => ruleset }

  source = "./submodules/ruleset"

  repository_name = var.repository_name
  ruleset        = each.value
}