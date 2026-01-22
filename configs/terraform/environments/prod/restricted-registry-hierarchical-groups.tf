# Layer 1 → Layer 2: Service Accounts join Hierarchical Groups

resource "google_cloud_identity_group_membership" "security_scanners_sa_to_hierarchical_group" {
  group = var.restricted_registry_hierarchical_groups.security_scanners_group_name

  preferred_member_key {
    id = google_service_account.kyma-security-scanners.email
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "markets_delivery_sa_to_hierarchical_group" {
  group = var.restricted_registry_hierarchical_groups.markets_delivery_group_name

  preferred_member_key {
    id = google_service_account.restricted-markets-artifactregistry-reader.email
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "image_builder_restricted_markets_sa_to_hierarchical_group" {
  group = var.restricted_registry_hierarchical_groups.image_builder_group_name

  preferred_member_key {
    id = google_service_account.kyma_project_image_builder_restricted_markets.email
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "image_builder_restricted_markets_sa_to_image_signer_group" {
  provider = google.kyma_project
  group    = var.restricted_registry_hierarchical_groups.image_signer_group_name

  preferred_member_key {
    id = google_service_account.kyma_project_image_builder_restricted_markets.email
  }

  roles {
    name = "MEMBER"
  }
}

# Layer 2 → Layer 3: Hierarchical Groups join Registry Access Groups

resource "google_cloud_identity_group_membership" "security_scanners_group_to_prod_read" {
  group = var.restricted_registry_iam_groups.prod_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.security_scanners
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "security_scanners_group_to_dev_read" {
  group = var.restricted_registry_iam_groups.dev_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.security_scanners
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "developers_group_to_prod_read" {
  group = var.restricted_registry_iam_groups.prod_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.developers
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "developers_group_to_dev_read" {
  group = var.restricted_registry_iam_groups.dev_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.developers
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "markets_delivery_group_to_prod_read" {
  group = var.restricted_registry_iam_groups.prod_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.markets_delivery
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "image_signer_group_to_prod_read" {
  group = var.restricted_registry_iam_groups.prod_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.image_signer
  }

  roles {
    name = "MEMBER"
  }
}

# Image builder with accesses to restricted registries
resource "google_cloud_identity_group_membership" "image_builder_group_to_dev_read" {
  group = var.restricted_registry_iam_groups.dev_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.image_builder
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "image_builder_group_to_dev_write" {
  group = var.restricted_registry_iam_groups.dev_write_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.image_builder
  }

  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "image_builder_group_to_prod_read" {
  group = var.restricted_registry_iam_groups.prod_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.image_builder
  }
  roles {
    name = "MEMBER"
  }
}

resource "google_cloud_identity_group_membership" "image_builder_group_to_prod_write" {
  group = var.restricted_registry_iam_groups.prod_write_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.image_builder
  }
  roles {
    name = "MEMBER"
  }
}