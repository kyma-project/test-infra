output "tekton_gatekeeper" {
  value     = module.tekton_gatekeeper
  sensitive = true
}

output "trusted_workload_gatekeeper" {
  value     = module.trusted_workload_gatekeeper
  sensitive = true
}

output "untrusted_workload_gatekeeper" {
  value     = module.untrusted_workload_gatekeeper
  sensitive = true
}
