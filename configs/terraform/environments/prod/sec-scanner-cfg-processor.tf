variable "sec-scanner-cfg-processor-service-account" {
  type = object({
    service-account-name        = string
    service-account-description = string
  })
  default = {
    service-account-name        = "sec-scanner-cfg-processor"
    service-account-description = "Identity of sec-scanner-cfg-processor"
  }
}

resource "google_service_account" "sec-scanner-cfg-processor" {
  account_id   = var.sec-scanner-cfg-processor-service-account.service-account-name
  display_name = var.sec-scanner-cfg-processor-service-account.service-account-name
  description  = var.sec-scanner-cfg-processor-service-account.service-account-description
}

removed {
  from = google_artifact_registry_repository_iam_member.sec_scanner_cfg_processor_kyma_modules_reader
  lifecycle { destroy = false }
}