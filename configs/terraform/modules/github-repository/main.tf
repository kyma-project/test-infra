terraform {
  required_providers {
    github = {
      source = "opentofu/github"
      version = "~> 6.6.0"
    }
  }
}
module "ruleset" {
  for_each = { for idx, ruleset in var.rulesets : idx => ruleset }

  source = "./submodules/ruleset"

  repository_name = var.repository_name
  ruleset        = each.value
}