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

resource "google_cloud_identity_group_membership" "developers_group_to_dev_write" {
  group = var.restricted_registry_iam_groups.dev_write_group_name

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

# Manager group for kyma-restricted-registry-developers to enable management through the CLI or web console.
resource "google_cloud_identity_group_membership" "kyma_developer_admin_as_developers_group_manager" {
  group = var.restricted_registry_hierarchical_groups.developers_group_name

  preferred_member_key {
    id = var.kyma_developer_admin_email
  }

  roles {
    name = "MANAGER"
  }
}

# ------------------------------------------------------------------------------
# Internal GitHub Team Members as Developers Group Owners
# ------------------------------------------------------------------------------
# Fetch the neighbors team from internal GitHub Enterprise instance and grant
# its members OWNER role on the restricted registry developers hierarchical group.
# This enables team members to manage group membership through CLI or web console.
#
# Note: github_team.members returns GitHub usernames (login names).
# We use github_user data source to fetch each user's email address from their
# GitHub profile, which should be their SAP email (firstname.lastname@sap.com).
# Users without a configured email in their GitHub profile are skipped.
# Service accounts are skipped - only users with usernames starting with I or D
# (I-numbers and D-numbers) are included.
# ------------------------------------------------------------------------------

# Data source to fetch the neighbors team from internal GitHub
data "github_team" "neighbors" {
  provider = github.internal_github
  slug     = var.internal_github_neighbors_team_slug
}

# Fetch user details for each team member to get their email address
data "github_user" "neighbors_team_members" {
  for_each = toset(data.github_team.neighbors.members)
  provider = github.internal_github
  username = each.value
}

# Local variable to filter out users without email addresses and service accounts
locals {
  neighbors_team_members_with_email = {
    for username, user in data.github_user.neighbors_team_members :
    username => user
    if(
      user.email != null &&
      user.email != "" &&
      (startswith(upper(username), "I") || startswith(upper(username), "D"))
    )
  }
}

# This user was already added as OWNER before Terraform management
import {
  id = "${var.restricted_registry_hierarchical_groups.developers_group_name}/memberships/patryk.dobrowolski@sap.com"
  to = google_cloud_identity_group_membership.neighbors_team_members_as_developers_group_owners["I583797"]
}

# This user was already added as OWNER before Terraform management
import {
  id = "${var.restricted_registry_hierarchical_groups.developers_group_name}/memberships/dawid.gala@sap.com"
  to = google_cloud_identity_group_membership.neighbors_team_members_as_developers_group_owners["I767604"]
}

# Grant each member of the neighbors team OWNER role on the developers hierarchical group
# Only users with a valid email address in their GitHub profile are included
resource "google_cloud_identity_group_membership" "neighbors_team_members_as_developers_group_owners" {
  for_each = local.neighbors_team_members_with_email

  group = var.restricted_registry_hierarchical_groups.developers_group_name

  preferred_member_key {
    id = each.value.email
  }

  roles {
    name = "OWNER"
  }
}