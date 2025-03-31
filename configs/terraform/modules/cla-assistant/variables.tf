variable "gcp_project_id" {
  type    = string
  default = "sap-kyma-prow"
}

variable "issue_comment_created_pubsub_topic_name" {
  type    = string
  default = "issue_comment.created"
}

variable "issue_comment_edited_pubsub_topic_name" {
  type    = string
  default = "issue_comment.edited"
}

variable "issue_comment_cloud_run_path" {
  type    = string
  default = "/issue-comment"
}
