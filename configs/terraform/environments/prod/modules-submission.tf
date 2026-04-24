removed {
  from = module.kyma_modules
  lifecycle {
    destroy = false
  }
}

removed {
  from = module.dev_kyma_modules
  lifecycle {
    destroy = false
  }
}

removed {
  from = google_service_account.kyma-submission-pipeline
  lifecycle {
    destroy = false
  }
}

removed {
  from = google_service_account.kyma_modules_reader
  lifecycle {
    destroy = false
  }
}

removed {
  from = google_artifact_registry_repository_iam_member.kyma_modules_registry_reader
  lifecycle {
    destroy = false
  }
}
