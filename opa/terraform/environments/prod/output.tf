output "tekton_gatekeeper" {
  value     = module.tekton_gatekeeper
  sensitive = true
}

output "trusted_workloads_gatekeeper" {
  value     = module.trusted_workloads_gatekeeper
  sensitive = true
}

output "untrusted_workloads_gatekeeper" {
  value     = module.untrusted_workloads_gatekeeper
  sensitive = true
}
