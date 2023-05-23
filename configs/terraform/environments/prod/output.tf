output "tekton_gatekeeper_constraints" {
  value     = module.tekton_gatekeeper_constraints
  sensitive = true
}

output "trusted_gatekeeper_constraints" {
  value     = module.trusted_workload_gatekeeper
  sensitive = true
}

output "untrusted_gatekeeper_constraints" {
  value     = module.untrusted_gatekeeper_constraints
  sensitive = true
}
