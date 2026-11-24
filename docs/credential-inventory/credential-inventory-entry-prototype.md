# Credential inventory — entry prototype (human- and AI-readable format)

> Status: **PROPOSAL FOR REVIEW**. This document defines the set of fields for a credential
> inventory entry. The machine-readable version (for scripts) is in
> `credential-inventory-entry-prototype.yaml`.

## Granularity principle

We inventory **per credential**, not per identity. A single identity may have multiple entries.
A separate entry is created for each of: GitHub User PAT, GCP service account key, GitHub App
credentials, database password, service account without keys, access through a RoleBinding, and
access through Workload Identity Federation.

## Entry fields

### Required fields

- `id` — stable, unique entry identifier (slug), one per credential.
- `credential_type` — kind of credential (value from the dictionary).
- `identity` — the identity the credential belongs to (full, unambiguous name).
- `identity_type` — type of identity (value from the dictionary).
- `identity_provider` — system that issues/manages the identity (`google-cloud`, `github`, `gardener`, `database`).
- `purpose` — usage context / purpose of the credential (value from the dictionary; e.g. `github-access`, `iaas-access`).
- `secret_store` — secret store used (value from the dictionary; `none` if there is no credential material).
- `storage_reference` — exact pointer within the store (secret name, path, resource).
- `used_by` — application or process the credential/identity is associated with.
- `used_where` — where it is used (list of execution environments).
- `owner` — owner / responsible team (consistent with the `owner` label in Secret Manager).
- `rotation_method` — rotation method (`automated`, `manual`, `tbd`, `not-applicable`).
- `rotation_frequency` — current rotation frequency (e.g. `90d`, `yearly`, `on-demand`, `unknown`).
- `last_rotation` — date of the last rotation (or `unknown`).
- `rotation_monitoring_source` — where the last rotation can be monitored/confirmed (value from the dictionary).
- `rotation_needs_customer_action` — whether rotation requires customer action (`yes`, `no`, `unknown`).
- `status` — entry state (`active`, `deprecated`, `to-rotate`, `unknown`).

### Optional / supplementary fields

- `tool` — tool (consistent with the `<tool>_<component>_<entity>` convention).
- `component` — tool component.
- `entity` — credential entity.
- `access_mechanism` — how access is realized when there is no material (`workload-identity-federation`, `role-binding`, `static-key`).
- `expires_at` — expiry date, if applicable.
- `scope_permissions` — scope of permissions / roles granted to the identity.
- `source_of_truth` — where the entry comes from (for audit and reproducibility).
- `last_reviewed` — date the entry was last reviewed.
- `notes` — free-form notes.

## Value dictionaries (proposal — to be extended)

### credential_type

- `github-user-pat` — GitHub User personal access token
- `github-app-key` — GitHub App private key
- `gcp-service-account-key` — GCP service account JSON key
- `gcp-service-account-no-key` — service account without keys (access through IAM)
- `database-password` — database password (e.g. Postgres)
- `kubernetes-rolebinding` — access through RoleBinding/ClusterRoleBinding
- `workload-identity-federation` — access through WIF (no credential material)
- `oidc-trust` — OIDC trust relationship (e.g. GitHub Actions → GCP)
- `static-token` — other static token/secret
- `ssh-key` — SSH key

### identity_type

- `github-user`
- `github-service-user` (service user / bot)
- `github-app`
- `gcp-service-account`
- `gcp-user`
- `kubernetes-service-account`
- `database-user`

### purpose

- `github-access` — access to GitHub (repos, API, Actions)
- `iaas-access` — access to infrastructure (Google Cloud, Gardener, etc.)
- `container-registry-access` — access to an image registry
- `image-signing` — signing images/artifacts
- `database-access` — database access
- `notification` — notification integrations (e.g. Slack)
- `ci-cd-automation` — CI/CD pipeline automation
- `monitoring-observability` — monitoring / observability

### secret_store

- `gcp-secret-manager`
- `vault-tools-sap` (vault.tools.sap)
- `sap-passvault` (SAP Passvault)
- `btp-credential-store` (BTP Credential Store service)
- `gardener-cluster-secret` (secret in a Gardener cluster)
- `github-secret` (repo / org / environment secret)
- `azure-devops-secret`
- `none` (no credential material — e.g. WIF, RoleBinding)

### rotation_method

- `automated`
- `manual`
- `tbd` (to be determined)
- `not-applicable`

### rotation_frequency (example values)

- `90d`, `180d`, `yearly` — fixed interval
- `on-demand` — rotation triggered manually as needed
- `not-applicable` — e.g. WIF/RoleBinding without material
- `unknown` — not established

### rotation_monitoring_source

- `gcp-secret-manager-versions` — secret versions / creation date in GCP Secret Manager
- `gcp-sa-key-metadata` — service account key metadata (`validAfterTime`) in GCP IAM
- `github-app-settings` — GitHub App settings → private keys
- `github-pat-settings` — GitHub user settings → tokens (expiry date)
- `vault-tools-sap-metadata` — secret metadata in vault.tools.sap
- `sap-passvault-metadata` — entry metadata in SAP Passvault
- `btp-credential-store-metadata` — entry metadata in BTP Credential Store
- `gardener-cluster-secret-metadata` — secret metadata in a Gardener cluster
- `azure-devops-secret-metadata` — secret/variable metadata in Azure DevOps
- `manual-tracking` — no automated source; tracked manually (e.g. in this inventory)
- `not-applicable` — no credential material (e.g. WIF, RoleBinding)

The value may be refined in parentheses, e.g. `gcp-secret-manager-versions (projects/sap-kyma-prow/secrets/...)`.

### rotation_needs_customer_action

- `yes` — rotation requires customer action
- `no` — rotation handled on our side / fully automated
- `unknown`

### used_where (example list item values)

- `gardener-cluster: <cluster-name>`
- `github-actions: <repo>/.github/workflows/<file>`
- `azure-devops: <pipeline>`
- `gcp-cloud-run: <service>`
- `local/manual`

## Example entries

### gcp-sa-key_image-builder_signer

- `id`: gcp-sa-key_image-builder_signer
- `credential_type`: gcp-service-account-key
- `identity`: image-builder-signer@sap-kyma-prow.iam.gserviceaccount.com
- `identity_type`: gcp-service-account
- `identity_provider`: google-cloud
- `purpose`: image-signing
- `secret_store`: gcp-secret-manager
- `storage_reference`: projects/sap-kyma-prow/secrets/image-builder_signer_gcp-sa-key
- `used_by`: image-builder
- `used_where`:
  - github-actions: kyma-project/test-infra/.github/workflows/image-builder.yml
- `owner`: neighbors
- `rotation_method`: manual
- `rotation_frequency`: on-demand
- `last_rotation`: unknown
- `rotation_monitoring_source`: gcp-secret-manager-versions (projects/sap-kyma-prow/secrets/image-builder_signer_gcp-sa-key)
- `rotation_needs_customer_action`: no
- `status`: active
- `tool`: image-builder
- `component`: signer
- `entity`: gcp-sa-key
- `scope_permissions`: roles/storage.objectAdmin
- `source_of_truth`: terraform: configs/terraform/environments/prod/service_accounts.tf
- `last_reviewed`: 2026-11-24

### wif_terraform-executor_gha

- `id`: wif_terraform-executor_gha
- `credential_type`: workload-identity-federation
- `identity`: terraform-executor@sap-kyma-prow.iam.gserviceaccount.com
- `identity_type`: gcp-service-account
- `identity_provider`: google-cloud
- `purpose`: iaas-access
- `secret_store`: none
- `storage_reference`: none
- `used_by`: terraform-executor
- `used_where`:
  - github-actions: kyma-project/test-infra (workflows using OIDC to GCP)
- `owner`: neighbors
- `rotation_method`: not-applicable
- `rotation_frequency`: not-applicable
- `last_rotation`: not-applicable
- `rotation_monitoring_source`: not-applicable
- `rotation_needs_customer_action`: no
- `status`: active
- `access_mechanism`: workload-identity-federation
- `source_of_truth`: terraform: configs/terraform/modules/gcp-workload-identity-federation
- `last_reviewed`: 2026-11-24

### github-app-key_kyma-bot_default

- `id`: github-app-key_kyma-bot_default
- `credential_type`: github-app-key
- `identity`: kyma-bot (GitHub App)
- `identity_type`: github-app
- `identity_provider`: github
- `purpose`: github-access
- `secret_store`: gcp-secret-manager
- `storage_reference`: projects/sap-kyma-prow/secrets/gha_default_github-kyma-app-key
- `used_by`: workflow-controller
- `used_where`:
  - github-actions: kyma-project/* (various workflows)
- `owner`: neighbors
- `rotation_method`: manual
- `rotation_frequency`: on-demand
- `last_rotation`: unknown
- `rotation_monitoring_source`: github-app-settings + gcp-secret-manager-versions (projects/sap-kyma-prow/secrets/gha_default_github-kyma-app-key)
- `rotation_needs_customer_action`: no
- `status`: active
- `last_reviewed`: 2026-11-24

### rolebinding_doc-collector_prod-cluster

- `id`: rolebinding_doc-collector_prod-cluster
- `credential_type`: kubernetes-rolebinding
- `identity`: doc-collector (KSA)
- `identity_type`: kubernetes-service-account
- `identity_provider`: gardener
- `purpose`: iaas-access
- `secret_store`: none
- `storage_reference`: none
- `used_by`: doc-collector
- `used_where`:
  - gardener-cluster: prod-cluster
- `owner`: neighbors
- `rotation_method`: not-applicable
- `rotation_frequency`: not-applicable
- `last_rotation`: not-applicable
- `rotation_monitoring_source`: not-applicable
- `rotation_needs_customer_action`: no
- `status`: active
- `access_mechanism`: role-binding
- `last_reviewed`: 2026-11-24

### db-password_dashboard_postgres

- `id`: db-password_dashboard_postgres
- `credential_type`: database-password
- `identity`: dashboard_app (database-user)
- `identity_type`: database-user
- `identity_provider`: database
- `purpose`: database-access
- `secret_store`: gardener-cluster-secret
- `storage_reference`: namespace/dashboard, secret/postgres-credentials, key/password
- `used_by`: dashboard
- `used_where`:
  - gardener-cluster: prod-cluster
- `owner`: neighbors
- `rotation_method`: manual
- `rotation_frequency`: unknown
- `last_rotation`: unknown
- `rotation_monitoring_source`: gardener-cluster-secret-metadata (namespace/dashboard, secret/postgres-credentials) — no dedicated monitoring
- `rotation_needs_customer_action`: no
- `status`: active
- `last_reviewed`: 2026-11-24
