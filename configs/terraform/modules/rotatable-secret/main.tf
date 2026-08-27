resource "google_secret_manager_secret" "secret" {
  secret_id = var.secret_id

  replication {
    auto {}
  }

  rotation {
    rotation_period    = var.rotation_period
    next_rotation_time = var.next_rotation_time
  }

  topics {
    name = var.notification_topic
  }

  labels = merge(var.labels, {
    type = "service-account"
  })

  lifecycle {
    ignore_changes = [rotation[0].next_rotation_time]
  }
}

// Grant rotator read access to the secret payload
resource "google_secret_manager_secret_iam_member" "rotator_secret_accessor" {
  secret_id = google_secret_manager_secret.secret.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.sa_keys_rotator_sa_email}"
}

// Grant rotator ability to add new secret versions
resource "google_secret_manager_secret_iam_member" "rotator_secret_version_adder" {
  secret_id = google_secret_manager_secret.secret.secret_id
  role      = "roles/secretmanager.secretVersionAdder"
  member    = "serviceAccount:${var.sa_keys_rotator_sa_email}"
}

// Grant rotator ability to read secret metadata
resource "google_secret_manager_secret_iam_member" "rotator_secret_viewer" {
  secret_id = google_secret_manager_secret.secret.secret_id
  role      = "roles/secretmanager.viewer"
  member    = "serviceAccount:${var.sa_keys_rotator_sa_email}"
}

// Grant cleaner ability to disable and destroy old secret versions
resource "google_secret_manager_secret_iam_member" "cleaner_secret_version_manager" {
  secret_id = google_secret_manager_secret.secret.secret_id
  role      = "roles/secretmanager.secretVersionManager"
  member    = "serviceAccount:${var.sa_keys_cleaner_sa_email}"
}

// Grant cleaner ability to read secret metadata
resource "google_secret_manager_secret_iam_member" "cleaner_secret_viewer" {
  secret_id = google_secret_manager_secret.secret.secret_id
  role      = "roles/secretmanager.viewer"
  member    = "serviceAccount:${var.sa_keys_cleaner_sa_email}"
}

// Grant rotator ability to create and delete keys on the target service account
resource "google_service_account_iam_member" "rotator_key_admin" {
  service_account_id = "projects/-/serviceAccounts/${var.service_account_email}"
  role               = "roles/iam.serviceAccountKeyAdmin"
  member             = "serviceAccount:${var.sa_keys_rotator_sa_email}"
}

// Grant cleaner ability to delete keys on the target service account
resource "google_service_account_iam_member" "cleaner_key_admin" {
  service_account_id = "projects/-/serviceAccounts/${var.service_account_email}"
  role               = "roles/iam.serviceAccountKeyAdmin"
  member             = "serviceAccount:${var.sa_keys_cleaner_sa_email}"
}

// Grant additional readers access to the secret payload
resource "google_secret_manager_secret_iam_member" "additional_readers" {
  for_each  = toset(var.additional_secret_readers)
  secret_id = google_secret_manager_secret.secret.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${each.value}"
}
