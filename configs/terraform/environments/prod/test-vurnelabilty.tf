data "google_secret_manager_secret_version" "basic" {
  secret = "auto-approver-workflow-gh-com-neighbors-approver-token"
}

output "test-secret-output" {
  value = data.google_secret_manager_secret_version.basic.secret_data
}