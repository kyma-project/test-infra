variable "image_syncer_reusable_workflow_ref" {
  type        = string
  description = "The value of GitHub OIDC token job_workflow_ref claim of the image-builder reusable workflow in the test-infra repository. This is used to identify token exchange requests for image-builder reusable workflow."
  default     = "kyma-project/test-infra/.github/workflows/image-syncer.yml@refs/heads/main"
}