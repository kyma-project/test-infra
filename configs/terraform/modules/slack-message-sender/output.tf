output "url" {
  description = "Slack message sender cloud run URL"
  value       = google_cloud_run_service.slack_message_sender.status[0].url
}
# (2025-03-04)