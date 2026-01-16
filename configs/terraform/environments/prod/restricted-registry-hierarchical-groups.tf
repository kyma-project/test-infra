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

resource "google_cloud_identity_group_membership" "markets_delivery_group_to_prod_read" {
  group = var.restricted_registry_iam_groups.prod_read_group_name

  preferred_member_key {
    id = var.restricted_registry_hierarchical_groups.markets_delivery
  }

  roles {
    name = "MEMBER"
  }
}
